package luaSandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type SandboxProfile uint8

const (
	HookSandbox SandboxProfile = iota
	InlineSandbox
)

type SandboxOptions struct {
	FSContext FSContext
	Profile   SandboxProfile
}

type Runtime struct {
	state   *lua.LState
	exits   map[*lua.LState]int
	cancels map[*lua.LState]context.CancelFunc
}

type Scope struct {
	runtime  *Runtime
	env      *lua.LTable
	readOnly map[*lua.LTable]readOnlyTable
}

type readOnlyTable struct {
	backing        *lua.LTable
	foldStringKeys bool
}

type Result struct {
	Exited   bool
	ExitCode int
}

// NewRuntime creates the shared Lua state used throughout one sklair build
func NewRuntime() *Runtime {
	L := lua.NewState(lua.Options{
		RegistrySize:        128,
		RegistryMaxSize:     512,
		SkipOpenLibs:        true,
		IncludeGoStackTrace: false,
		//MinimizeStackMemory: false,
	})
	runtime := &Runtime{
		state:   L,
		exits:   make(map[*lua.LState]int),
		cancels: make(map[*lua.LState]context.CancelFunc),
	}
	L.SetContext(context.Background())

	OpenSandboxedDefault(L, runtime)
	OpenSandboxedCustom(L)

	return runtime
}

func (r *Runtime) Close() {
	r.state.Close()
}

func (r *Runtime) exit(thread *lua.LState, code int) {
	r.exits[thread] = code
	if cancel := r.cancels[thread]; cancel != nil {
		cancel()
	}
}

// NewScope creates an isolated global environment which inherits the shared
// libraries and (future) module cache from the build's root Lua state
func (r *Runtime) NewScope(options SandboxOptions) *Scope {
	env := r.state.NewTable()
	metatable := r.state.NewTable()
	metatable.RawSetString("__index", r.state.Env)
	metatable.RawSetString("__metatable", lua.LString("Sklair scope"))
	r.state.SetMetatable(env, metatable)
	env.RawSetString("_G", env)
	for _, name := range []string{"table", "os", "string", "math", "json"} {
		if shared, ok := r.state.GetGlobal(name).(*lua.LTable); ok {
			env.RawSetString(name, cloneTable(r.state, shared))
		}
	}
	env.RawSetString("fs", newFsModule(r.state, &options))

	return &Scope{runtime: r, env: env, readOnly: make(map[*lua.LTable]readOnlyTable)}
}

func cloneTable(L *lua.LState, source *lua.LTable) *lua.LTable {
	clone := L.NewTable()
	source.ForEach(func(key lua.LValue, value lua.LValue) {
		clone.RawSet(key, value)
	})
	clone.Metatable = source.Metatable
	return clone
}

func (s *Scope) SetModule(name string, functions map[string]lua.LGFunction) {
	module := s.runtime.state.NewTable()
	s.runtime.state.SetFuncs(module, functions)
	s.env.RawSetString(name, module)
}

func (s *Scope) SetReadOnly(name string, values map[string]any) error {
	backing := s.runtime.state.NewTable()
	keys := sortedKeys(values)
	for _, key := range keys {
		value := values[key]
		converted, err := s.readOnlyValue(name+"."+key, value)
		if err != nil {
			return err
		}
		backing.RawSetString(strings.ToLower(key), converted)
	}
	s.setReadOnly(name, backing, true)
	return nil
}

func (s *Scope) readOnlyValue(name string, value any) (lua.LValue, error) {
	switch value := value.(type) {
	case nil:
		return lua.LNil, nil
	case bool:
		return lua.LBool(value), nil
	case string:
		return lua.LString(value), nil
	case float64:
		return lua.LNumber(value), nil
	case []any:
		backing := s.runtime.state.NewTable()
		for index, child := range value {
			converted, err := s.readOnlyValue(fmt.Sprintf("%s[%d]", name, index+1), child)
			if err != nil {
				return nil, err
			}
			backing.RawSetInt(index+1, converted)
		}
		return s.readOnlyProxy(name, backing, false), nil
	case map[string]any:
		backing := s.runtime.state.NewTable()
		keys := sortedKeys(value)
		for _, key := range keys {
			child := value[key]
			converted, err := s.readOnlyValue(name+"."+key, child)
			if err != nil {
				return nil, err
			}
			backing.RawSetString(key, converted)
		}
		return s.readOnlyProxy(name, backing, false), nil
	default:
		return nil, fmt.Errorf("cannot expose %T as a read-only Lua value", value)
	}
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *Scope) setReadOnly(name string, backing *lua.LTable, foldStringKeys bool) {
	proxy := s.readOnlyProxy(name, backing, foldStringKeys)
	s.env.RawSetString(name, proxy)
	s.protectReadOnly(name)
}

func (s *Scope) readOnlyProxy(name string, backing *lua.LTable, foldStringKeys bool) *lua.LTable {
	proxy := s.runtime.state.NewTable()
	metatable := s.runtime.state.NewTable()
	metatable.RawSetString("__index", s.runtime.state.NewFunction(func(L *lua.LState) int {
		key := L.Get(2)
		if foldStringKeys {
			if stringKey, ok := key.(lua.LString); ok {
				key = lua.LString(strings.ToLower(string(stringKey)))
			}
		}
		L.Push(backing.RawGet(key))
		return 1
	}))
	metatable.RawSetString("__newindex", s.runtime.state.NewFunction(func(L *lua.LState) int {
		L.RaiseError("%s is read-only", name)
		return 0
	}))
	metatable.RawSetString("__len", s.runtime.state.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(backing.Len()))
		return 1
	}))
	metatable.RawSetString("__metatable", lua.LString("read-only "+name))
	s.runtime.state.SetMetatable(proxy, metatable)
	s.readOnly[proxy] = readOnlyTable{backing: backing, foldStringKeys: foldStringKeys}
	return proxy
}

func (s *Scope) protectReadOnly(name string) {
	s.env.RawSetString("next", s.runtime.state.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		if readOnly, exists := s.readOnly[table]; exists {
			table = readOnly.backing
		}
		key := lua.LNil
		if L.GetTop() > 1 {
			key = L.Get(2)
		}
		key, value := table.Next(key)
		if key == lua.LNil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(key)
		L.Push(value)
		return 2
	}))
	s.env.RawSetString("ipairs", s.runtime.state.NewFunction(func(L *lua.LState) int {
		L.CheckTable(1)
		iterator := s.runtime.state.NewFunction(func(L *lua.LState) int {
			table := L.CheckTable(1)
			index := L.CheckInt(2) + 1
			if readOnly, exists := s.readOnly[table]; exists {
				table = readOnly.backing
			}
			value := table.RawGetInt(index)
			if value == lua.LNil {
				return 0
			}
			L.Push(lua.LNumber(index))
			L.Push(value)
			return 2
		})
		L.Push(iterator)
		L.Push(L.Get(1))
		L.Push(lua.LNumber(0))
		return 3
	}))
	s.env.RawSetString("pairs", s.runtime.state.NewFunction(func(L *lua.LState) int {
		L.CheckTable(1)
		L.Push(s.env.RawGetString("next"))
		L.Push(L.Get(1))
		L.Push(lua.LNil)
		return 3
	}))
	s.env.RawSetString("rawget", s.runtime.state.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		key := L.Get(2)
		if readOnly, exists := s.readOnly[table]; exists {
			if readOnly.foldStringKeys {
				if stringKey, ok := key.(lua.LString); ok {
					key = lua.LString(strings.ToLower(string(stringKey)))
				}
			}
			L.Push(readOnly.backing.RawGet(key))
			return 1
		}
		L.Push(table.RawGet(key))
		return 1
	}))
	s.env.RawSetString("rawset", s.runtime.state.NewFunction(func(L *lua.LState) int {
		table := L.CheckTable(1)
		key := L.CheckAny(2)
		if _, readOnly := s.readOnly[table]; readOnly || table == s.env && key == lua.LString(name) {
			L.RaiseError("%s is read-only", name)
			return 0
		}
		table.RawSet(key, L.CheckAny(3))
		L.Push(table)
		return 1
	}))
	if table, ok := s.env.RawGetString("table").(*lua.LTable); ok {
		for _, functionName := range []string{"insert", "remove", "sort"} {
			original, ok := table.RawGetString(functionName).(*lua.LFunction)
			if !ok {
				continue
			}
			table.RawSetString(functionName, s.runtime.state.NewFunction(func(L *lua.LState) int {
				if target, ok := L.Get(1).(*lua.LTable); ok {
					if _, readOnly := s.readOnly[target]; readOnly {
						L.RaiseError("%s is read-only", name)
						return 0
					}
				}
				argumentCount := L.GetTop()
				arguments := make([]lua.LValue, argumentCount)
				for i := range arguments {
					arguments[i] = L.Get(i + 1)
				}
				if err := L.CallByParam(lua.P{Fn: original, NRet: lua.MultRet, Protect: true}, arguments...); err != nil {
					L.RaiseError(err.Error())
					return 0
				}
				return L.GetTop() - argumentCount
			}))
		}
	}
}

// Run executes one Lua usage on a fresh child thread. The scope environment
// survives between usages, while stacks and os.exit are confined to this lua usage / run
func (s *Scope) Run(source io.Reader, name string) (Result, error) {
	function, err := s.runtime.state.Load(source, name)
	if err != nil {
		return Result{}, err
	}
	return s.run(function.Proto)
}

func (s *Scope) RunPrototype(prototype *lua.FunctionProto) (Result, error) {
	return s.run(prototype)
}

func (s *Scope) run(prototype *lua.FunctionProto) (Result, error) {
	thread, cancel := s.runtime.state.NewThread()
	thread.Env = s.env
	s.runtime.cancels[thread] = cancel
	defer func() {
		delete(s.runtime.cancels, thread)
		cancel()
	}()

	function := thread.NewFunctionFromProto(prototype)
	state, err, _ := s.runtime.state.Resume(thread, function)
	if code, exited := s.runtime.exits[thread]; exited {
		delete(s.runtime.exits, thread)
		return Result{Exited: true, ExitCode: code}, nil
	}
	if err != nil {
		return Result{}, err
	}
	if state == lua.ResumeYield {
		return Result{}, errors.New("unexpected Lua thread yield")
	}
	return Result{}, nil
}
