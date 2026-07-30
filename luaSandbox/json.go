package luaSandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	lua "github.com/yuin/gopher-lua"
)

type jsonNull struct{}

func (s *Scope) openJSON() {
	L := s.runtime.state
	s.jsonNull = L.NewUserData()
	s.jsonNull.Value = jsonNull{}

	module := L.NewTable()
	L.SetFuncs(module, map[string]lua.LGFunction{
		"array":  s.jsonTableConstructor(arrayTable),
		"decode": s.decodeJSON,
		"encode": s.encodeJSON,
		"object": s.jsonTableConstructor(objectTable),
	})
	module.RawSetString("null", s.jsonNull)
	s.env.RawSetString("json", module)
}

func (s *Scope) jsonTableConstructor(kind tableKind) lua.LGFunction {
	return func(L *lua.LState) int {
		if L.GetTop() > 1 {
			L.RaiseError("json.%s expects zero arguments or one table", kind.name())
			return 0
		}
		table := L.NewTable()
		if L.GetTop() == 1 {
			table = L.CheckTable(1)
		}
		s.tableKinds[table] = kind
		L.Push(table)
		return 1
	}
}

func (k tableKind) name() string {
	if k == arrayTable {
		return "array"
	}
	return "object"
}

func (s *Scope) decodeJSON(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.RaiseError("json.decode expects exactly one string")
		return 0
	}
	value, err := parseJSON(L.CheckString(1))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	converted, err := s.luaStructuredValue(value)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(converted)
	return 1
}

func (s *Scope) encodeJSON(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.RaiseError("json.encode expects exactly one value")
		return 0
	}
	value, err := s.StructuredValue(L.Get(1), "value")
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(encoded))
	return 1
}

func parseJSON(source string) (any, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(source))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("JSON contains more than one value")
		}
		return nil, err
	}
	return value, nil
}
