package luaSandbox

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

type tableKind uint8

const (
	unknownTable tableKind = iota
	arrayTable
	objectTable
)

// StructuredValue snapshots a Lua value into the JSON-like value model shared
// by Sklair's Lua libraries and generated document props
func (s *Scope) StructuredValue(value lua.LValue, path string) (any, error) {
	return s.structuredValue(value, path, make(map[*lua.LTable]string))
}

func (s *Scope) structuredValue(value lua.LValue, path string, active map[*lua.LTable]string) (any, error) {
	switch value := value.(type) {
	case *lua.LNilType:
		return nil, nil
	case lua.LBool:
		return bool(value), nil
	case lua.LString:
		return string(value), nil
	case lua.LNumber:
		number := float64(value)
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return nil, fmt.Errorf("%s contains a non-finite number", path)
		}
		return number, nil
	case *lua.LTable:
		return s.structuredTable(value, path, active)
	case *lua.LUserData:
		if value == s.jsonNull {
			return nil, nil
		}
		return nil, fmt.Errorf("%s contains unsupported userdata", path)
	default:
		return nil, fmt.Errorf("%s contains unsupported %s", path, value.Type().String())
	}
}

func (s *Scope) structuredTable(table *lua.LTable, path string, active map[*lua.LTable]string) (any, error) {
	if readOnly, exists := s.readOnly[table]; exists {
		table = readOnly.backing
	}
	if previous, cyclic := active[table]; cyclic {
		return nil, fmt.Errorf("%s contains a cycle back to %s", path, previous)
	}
	active[table] = path
	defer delete(active, table)

	numeric := make(map[int]lua.LValue)
	object := make(map[string]lua.LValue)
	var keyError error
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if keyError != nil {
			return
		}
		switch key := key.(type) {
		case lua.LString:
			object[string(key)] = value
		case lua.LNumber:
			number := float64(key)
			if number < 1 || math.Trunc(number) != number || number > float64(math.MaxInt) {
				keyError = fmt.Errorf("%s has an invalid array key %s", path, key.String())
				return
			}
			numeric[int(number)] = value
		default:
			keyError = fmt.Errorf("%s has an unsupported %s key", path, key.Type().String())
		}
	})
	if keyError != nil {
		return nil, keyError
	}

	kind := s.tableKinds[table]
	if kind == unknownTable {
		switch {
		case len(numeric) > 0 && len(object) > 0:
			return nil, fmt.Errorf("%s mixes array and object keys", path)
		case len(numeric) > 0:
			kind = arrayTable
		default:
			kind = objectTable
		}
	}

	switch kind {
	case arrayTable:
		if len(object) > 0 {
			return nil, fmt.Errorf("%s is marked as an array but contains object keys", path)
		}
		array := make([]any, len(numeric))
		for index := 1; index <= len(numeric); index++ {
			child, exists := numeric[index]
			if !exists {
				return nil, fmt.Errorf("%s is a sparse array; missing index %d", path, index)
			}
			converted, err := s.structuredValue(child, fmt.Sprintf("%s[%d]", path, index), active)
			if err != nil {
				return nil, err
			}
			array[index-1] = converted
		}
		return array, nil
	case objectTable:
		if len(numeric) > 0 {
			return nil, fmt.Errorf("%s is marked as an object but contains array keys", path)
		}
		keys := make([]string, 0, len(object))
		for key := range object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		converted := make(map[string]any, len(object))
		for _, key := range keys {
			child, err := s.structuredValue(object[key], path+"."+key, active)
			if err != nil {
				return nil, err
			}
			converted[key] = child
		}
		return converted, nil
	default:
		return nil, fmt.Errorf("%s has an unknown table kind", path)
	}
}

func (s *Scope) luaStructuredValue(value any) (lua.LValue, error) {
	switch value := value.(type) {
	case nil:
		return s.jsonNull, nil
	case bool:
		return lua.LBool(value), nil
	case string:
		return lua.LString(value), nil
	case float64:
		return lua.LNumber(value), nil
	case json.Number:
		number, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON number %q", value)
		}
		return lua.LNumber(number), nil
	case []any:
		table := s.runtime.state.CreateTable(len(value), 0)
		s.tableKinds[table] = arrayTable
		for _, child := range value {
			converted, err := s.luaStructuredValue(child)
			if err != nil {
				return nil, err
			}
			table.Append(converted)
		}
		return table, nil
	case map[string]any:
		table := s.runtime.state.CreateTable(0, len(value))
		s.tableKinds[table] = objectTable
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			converted, err := s.luaStructuredValue(value[key])
			if err != nil {
				return nil, err
			}
			table.RawSetString(key, converted)
		}
		return table, nil
	default:
		return nil, fmt.Errorf("cannot expose %T as a structured Lua value", value)
	}
}
