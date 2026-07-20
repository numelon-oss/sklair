package building

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"sklair/luaSandbox"

	lua "github.com/yuin/gopher-lua"
)

var renderFields = map[string]struct{}{
	"layout": {},
	"output": {},
	"source": {},
	"props":  {},
	"body":   {},
}

func openDocuments(scope *luaSandbox.Scope, queue *documentQueue, hook string) {
	arrays := make(map[*lua.LTable]struct{})
	scope.SetModule("documents", map[string]lua.LGFunction{
		"array":  arrayConstructor(arrays),
		"render": renderDocument(queue, arrays, hook),
	})
}

func arrayConstructor(arrays map[*lua.LTable]struct{}) lua.LGFunction {
	return func(L *lua.LState) int {
		if L.GetTop() > 1 {
			L.RaiseError("documents.array expects zero arguments or one table")
			return 0
		}
		var array *lua.LTable
		if L.GetTop() == 1 {
			array = L.CheckTable(1)
		} else {
			array = L.NewTable()
		}
		arrays[array] = struct{}{}
		L.Push(array)
		return 1
	}
}

func renderDocument(queue *documentQueue, arrays map[*lua.LTable]struct{}, hook string) lua.LGFunction {
	return func(L *lua.LState) int {
		if L.GetTop() != 1 {
			L.RaiseError("documents.render expects exactly one request table")
			return 0
		}
		requestTable := L.CheckTable(1)
		if err := validateRenderFields(requestTable); err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}

		layout, err := requiredRenderString(requestTable, "layout")
		if err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		output, err := requiredRenderString(requestTable, "output")
		if err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		source, err := optionalRenderString(requestTable, "source")
		if err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		body, err := optionalRenderString(requestTable, "body")
		if err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}

		props, err := snapshotProps(requestTable.RawGetString("props"), arrays)
		if err != nil {
			L.RaiseError("invalid documents.render props : %s", err.Error())
			return 0
		}
		if err := queue.add(renderRequest{
			layout: layout,
			output: output,
			source: source,
			body:   body,
			props:  props,
			hook:   hook,
		}); err != nil {
			L.RaiseError("%s", err.Error())
			return 0
		}
		return 0
	}
}

func validateRenderFields(request *lua.LTable) error {
	invalid := make([]string, 0)
	nonString := false
	request.ForEach(func(key lua.LValue, _ lua.LValue) {
		name, ok := key.(lua.LString)
		if !ok {
			nonString = true
			return
		}
		if _, allowed := renderFields[string(name)]; !allowed {
			invalid = append(invalid, string(name))
		}
	})
	if nonString {
		return fmt.Errorf("documents.render field names must be strings")
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		return fmt.Errorf("unknown documents.render field %q", invalid[0])
	}
	return nil
}

func requiredRenderString(request *lua.LTable, name string) (string, error) {
	value := request.RawGetString(name)
	if value == lua.LNil {
		return "", fmt.Errorf("documents.render requires %q", name)
	}
	stringValue, ok := value.(lua.LString)
	if !ok {
		return "", fmt.Errorf("documents.render field %q must be a string", name)
	}
	if strings.TrimSpace(string(stringValue)) == "" {
		return "", fmt.Errorf("documents.render field %q cannot be empty", name)
	}
	return string(stringValue), nil
}

func optionalRenderString(request *lua.LTable, name string) (string, error) {
	value := request.RawGetString(name)
	if value == lua.LNil {
		return "", nil
	}
	stringValue, ok := value.(lua.LString)
	if !ok {
		return "", fmt.Errorf("documents.render field %q must be a string or nil", name)
	}
	if name == "source" && strings.TrimSpace(string(stringValue)) == "" {
		return "", fmt.Errorf("documents.render field %q cannot be empty", name)
	}
	return string(stringValue), nil
}

func snapshotProps(value lua.LValue, arrays map[*lua.LTable]struct{}) (map[string]sklairValue, error) {
	if value == lua.LNil {
		return map[string]sklairValue{}, nil
	}
	table, ok := value.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("expected a table or nil, got %s", value.Type().String())
	}
	snapshot, err := snapshotLuaTable(table, "props", arrays, make(map[*lua.LTable]string))
	if err != nil {
		return nil, err
	}
	if snapshot.kind != objectValue {
		return nil, fmt.Errorf("props must be an object")
	}

	props := make(map[string]sklairValue, len(snapshot.object))
	for suppliedName, value := range snapshot.object {
		name, err := propName(suppliedName)
		if err != nil {
			return nil, err
		}
		if _, exists := props[name]; exists {
			return nil, fmt.Errorf("prop %q is supplied more than once", name)
		}
		props[name] = value
	}
	return props, nil
}

func snapshotLua(value lua.LValue, path string, arrays map[*lua.LTable]struct{}, active map[*lua.LTable]string) (sklairValue, error) {
	switch value := value.(type) {
	case *lua.LNilType:
		return sklairValue{kind: nilValue}, nil
	case lua.LBool:
		return sklairValue{kind: booleanValue, boolean: bool(value)}, nil
	case lua.LString:
		return sklairValue{kind: stringValue, string: string(value)}, nil
	case lua.LNumber:
		number := float64(value)
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return sklairValue{}, fmt.Errorf("%s contains a non-finite number", path)
		}
		return sklairValue{kind: numberValue, number: number}, nil
	case *lua.LTable:
		return snapshotLuaTable(value, path, arrays, active)
	default:
		return sklairValue{}, fmt.Errorf("%s contains unsupported %s", path, value.Type().String())
	}
}

func snapshotLuaTable(table *lua.LTable, valuePath string, arrays map[*lua.LTable]struct{}, active map[*lua.LTable]string) (sklairValue, error) {
	if previous, cyclic := active[table]; cyclic {
		return sklairValue{}, fmt.Errorf("%s contains a cycle back to %s", valuePath, previous)
	}
	active[table] = valuePath
	defer delete(active, table)

	numeric := make(map[int]lua.LValue)
	object := make(map[string]lua.LValue)
	invalidKey := false
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		switch key := key.(type) {
		case lua.LString:
			object[string(key)] = value
		case lua.LNumber:
			number := float64(key)
			if number < 1 || math.Trunc(number) != number || number > float64(math.MaxInt) {
				invalidKey = true
				return
			}
			numeric[int(number)] = value
		default:
			invalidKey = true
		}
	})
	if invalidKey {
		return sklairValue{}, fmt.Errorf("%s has a non-string or invalid array key", valuePath)
	}
	if len(numeric) > 0 && len(object) > 0 {
		return sklairValue{}, fmt.Errorf("%s mixes array and object keys", valuePath)
	}

	_, explicitlyArray := arrays[table]
	if len(numeric) > 0 || explicitlyArray {
		if len(object) > 0 {
			return sklairValue{}, fmt.Errorf("%s is marked as an array but contains object keys", valuePath)
		}
		array := make([]sklairValue, len(numeric))
		for index := 1; index <= len(numeric); index++ {
			child, exists := numeric[index]
			if !exists {
				return sklairValue{}, fmt.Errorf("%s is a sparse array; missing index %d", valuePath, index)
			}
			snapshot, err := snapshotLua(child, fmt.Sprintf("%s[%d]", valuePath, index), arrays, active)
			if err != nil {
				return sklairValue{}, err
			}
			array[index-1] = snapshot
		}
		return sklairValue{kind: arrayValue, array: array}, nil
	}

	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make(map[string]sklairValue, len(object))
	for _, key := range keys {
		snapshot, err := snapshotLua(object[key], valuePath+"."+key, arrays, active)
		if err != nil {
			return sklairValue{}, err
		}
		values[key] = snapshot
	}
	return sklairValue{kind: objectValue, object: values}, nil
}
