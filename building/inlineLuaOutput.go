package building

import (
	"bytes"
	"errors"
	"fmt"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/luaSandbox"
	"sort"
	"strings"
	"unicode"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/net/html"
)

type luaEmitter struct {
	parent *html.Node
	nodes  []*html.Node
}

func openSklair(scope *luaSandbox.Scope, emitter *luaEmitter, compiler *staticLuaCompiler) {
	scope.SetModule("sklair", map[string]lua.LGFunction{
		"text":      emitText(emitter),
		"html":      emitHTML(emitter),
		"component": emitComponent(emitter, compiler),
	})
}

func emitText(emitter *luaEmitter) lua.LGFunction {
	return func(L *lua.LState) int {
		value := L.CheckAny(1)
		switch value.Type() {
		case lua.LTString, lua.LTNumber, lua.LTBool:
			emitter.nodes = append(emitter.nodes, &html.Node{Type: html.TextNode, Data: value.String()})
		default:
			L.RaiseError("sklair.text expects a string, number, or boolean")
		}
		return 0
	}
}

func emitHTML(emitter *luaEmitter) lua.LGFunction {
	return func(L *lua.LState) int {
		source := L.CheckString(1)
		nodes, err := html.ParseFragment(bytes.NewBufferString(source), emitter.parent)
		if err != nil {
			L.RaiseError("could not parse emitted HTML: %s", err.Error())
			return 0
		}
		emitter.nodes = append(emitter.nodes, nodes...)
		return 0
	}
}

func emitComponent(emitter *luaEmitter, compiler *staticLuaCompiler) lua.LGFunction {
	return func(L *lua.LState) int {
		name := strings.ToLower(strings.TrimSpace(L.CheckString(1)))
		if name == "" || strings.ContainsAny(name, " \t\r\n") {
			L.RaiseError("sklair.component expects a non-empty component name without whitespace")
			return 0
		}
		if htmlUtilities.HtmlTags[name] {
			L.RaiseError("sklair.component cannot invoke native HTML element %q", name)
			return 0
		}
		if _, exists := compiler.components[name]; !exists {
			L.RaiseError("component %q does not exist", name)
			return 0
		}
		if insideTag(emitter.parent, "head") {
			L.RaiseError("components cannot be emitted into <head>")
			return 0
		}

		attributes, err := luaComponentProps(L, 2)
		if err != nil {
			L.RaiseError("invalid props for component %s: %s", name, err.Error())
			return 0
		}
		if _, runtimeTemplate := compiler.templates[name]; runtimeTemplate && len(attributes) > 0 {
			L.RaiseError("runtime template component %s registration cannot have props", name)
			return 0
		}

		emitter.nodes = append(emitter.nodes, &html.Node{
			Type: html.ElementNode,
			Data: name,
			Attr: attributes,
		})
		return 0
	}
}

func luaComponentProps(L *lua.LState, index int) ([]html.Attribute, error) {
	if L.GetTop() < index || L.Get(index) == lua.LNil {
		return nil, nil
	}
	table, ok := L.Get(index).(*lua.LTable)
	if !ok {
		return nil, errors.New("props must be a table")
	}

	values := make(map[string]html.Attribute)
	seen := make(map[string]struct{})
	var propErr error
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if propErr != nil {
			return
		}
		if key.Type() != lua.LTString {
			propErr = errors.New("prop names must be strings")
			return
		}
		name := strings.ToLower(strings.TrimSpace(key.String()))
		if err := validateAttributeName(name); err != nil {
			propErr = err
			return
		}
		if _, duplicate := seen[name]; duplicate {
			propErr = fmt.Errorf("prop %q is supplied more than once", name)
			return
		}
		seen[name] = struct{}{}

		switch value.Type() {
		case lua.LTString, lua.LTNumber:
			values[name] = html.Attribute{Key: name, Val: value.String()}
		case lua.LTBool:
			if lua.LVAsBool(value) {
				values[name] = html.Attribute{Key: name}
			}
		case lua.LTNil:
		default:
			propErr = fmt.Errorf("prop %q must be a string, number, boolean, or nil", name)
		}
	})
	if propErr != nil {
		return nil, propErr
	}

	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	attributes := make([]html.Attribute, 0, len(names))
	for _, name := range names {
		attributes = append(attributes, values[name])
	}
	return attributes, nil
}

func validateAttributeName(name string) error {
	if name == "" {
		return errors.New("prop name cannot be empty")
	}
	for _, character := range name {
		if unicode.IsSpace(character) || unicode.IsControl(character) || strings.ContainsRune("\"'/>=", character) {
			return fmt.Errorf("invalid prop name %q", name)
		}
	}
	return nil
}

func validateLuaOutput(nodes []*html.Node, parent *html.Node, components map[string]discovery.ComponentSource) error {
	for _, node := range nodes {
		if node.Type == html.ElementNode && node.Data == "script" {
			matched, _, err := classifyInlineLua(node)
			if err != nil {
				return err
			}
			if matched {
				return errors.New("Lua emitted by Lua is not allowed")
			}
		}
		if insideTag(parent, "head") && node.Type == html.ElementNode {
			if _, component := components[strings.ToLower(node.Data)]; component {
				return fmt.Errorf("component %s cannot be emitted into <head>", node.Data)
			}
		}
		if err := validateLuaOutput(htmlUtilities.GetAllChildren(node), parent, components); err != nil {
			return err
		}
	}
	return nil
}

func insideTag(node *html.Node, tag string) bool {
	for current := node; current != nil; current = current.Parent {
		if current.Type == html.ElementNode && current.Data == tag {
			return true
		}
	}
	return false
}
