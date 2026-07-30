package building

import (
	"fmt"
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
	scope.SetModule("documents", map[string]lua.LGFunction{
		"render": renderDocument(scope, queue, hook),
	})
}

func renderDocument(scope *luaSandbox.Scope, queue *documentQueue, hook string) lua.LGFunction {
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

		props, err := snapshotProps(scope, requestTable.RawGetString("props"))
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

func snapshotProps(scope *luaSandbox.Scope, value lua.LValue) (map[string]sklairValue, error) {
	if value == lua.LNil {
		return map[string]sklairValue{}, nil
	}
	if _, ok := value.(*lua.LTable); !ok {
		return nil, fmt.Errorf("expected a table or nil, got %s", value.Type().String())
	}
	snapshot, err := scope.StructuredValue(value, "props")
	if err != nil {
		return nil, err
	}
	object, ok := snapshot.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("props must be an object")
	}

	props := make(map[string]sklairValue, len(object))
	for suppliedName, value := range object {
		name, err := propName(suppliedName)
		if err != nil {
			return nil, err
		}
		if _, exists := props[name]; exists {
			return nil, fmt.Errorf("prop %q is supplied more than once", name)
		}
		props[name] = structuredSklairValue(value)
	}
	return props, nil
}
