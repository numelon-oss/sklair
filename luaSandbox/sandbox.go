package luaSandbox

import (
	"context"
	"errors"
	"io"

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
	runtime *Runtime
	env     *lua.LTable
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
	r.state.SetMetatable(env, metatable)
	env.RawSetString("_G", env)
	for _, name := range []string{"table", "os", "string", "math", "json"} {
		if shared, ok := r.state.GetGlobal(name).(*lua.LTable); ok {
			env.RawSetString(name, cloneTable(r.state, shared))
		}
	}
	env.RawSetString("fs", newFsModule(r.state, &options))

	return &Scope{runtime: r, env: env}
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

// Run executes one Lua usage on a fresh child thread. The scope environment
// survives between usages, while stacks and os.exit are confined to this lua usage / run
func (s *Scope) Run(source io.Reader, name string) (Result, error) {
	thread, cancel := s.runtime.state.NewThread()
	thread.Env = s.env
	s.runtime.cancels[thread] = cancel
	defer func() {
		delete(s.runtime.cancels, thread)
		cancel()
	}()

	function, err := thread.Load(source, name)
	if err != nil {
		return Result{}, err
	}
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
