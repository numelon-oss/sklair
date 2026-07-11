package building

import (
	"errors"
	"fmt"
	"sklair/discovery"
	"sklair/luaSandbox"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const inlineLuaType = "application/x-sklair-lua"
const dynamicLuaMarker = "sklair-dynamic-lua"

type staticLuaCompiler struct {
	runtime    *luaSandbox.Runtime
	fsContext  luaSandbox.FSContext
	components map[string]discovery.ComponentSource
	templates  map[string]struct{}
}

type inlineLuaScript struct {
	node   *html.Node
	static bool
}

func (c *staticLuaCompiler) prepare(root *html.Node, source string, runtimeTemplate bool) (*dynamicLuaDefinition, error) {
	scripts, err := collectInlineLua(root)
	if err != nil {
		return nil, err
	}
	if len(scripts) == 0 {
		return nil, nil
	}
	if runtimeTemplate {
		return nil, errors.New("runtime template components cannot contain compile-time Lua")
	}

	for i, script := range scripts {
		if script.static && beneathSklairIf(script.node) {
			return nil, fmt.Errorf("static Lua block %d cannot be nested beneath sklair-if", i+1)
		}
	}

	dynamic := &dynamicLuaDefinition{props: make(map[string]struct{})}
	for i, script := range scripts {
		if script.static {
			continue
		}
		block := i + 1
		chunkName := fmt.Sprintf("%s:dynamic-lua-%d", source, block)
		prepared, props, open, err := prepareDynamicLua(scriptSource(script.node), chunkName, block)
		if err != nil {
			return nil, fmt.Errorf("dynamic Lua block %d in %s failed\n%s", block, source, err.Error())
		}
		index := len(dynamic.blocks)
		dynamic.blocks = append(dynamic.blocks, prepared)
		for name := range props {
			dynamic.props[name] = struct{}{}
		}
		dynamic.open = dynamic.open || open
		marker := &html.Node{Type: html.ElementNode, Data: dynamicLuaMarker, Attr: []html.Attribute{{Key: "block", Val: strconv.Itoa(index)}}}
		script.node.Parent.InsertBefore(marker, script.node)
		script.node.Parent.RemoveChild(script.node)
	}

	emitter := &luaEmitter{}
	scope := c.runtime.NewScope(luaSandbox.SandboxOptions{
		Profile:   luaSandbox.InlineSandbox,
		FSContext: c.fsContext,
	})
	openSklair(scope, emitter, c)

	for i, script := range scripts {
		if !script.static {
			continue
		}
		emitter.parent = script.node.Parent
		emitter.nodes = nil
		block := i + 1
		chunkName := fmt.Sprintf("%s:static-lua-%d", source, block)
		result, err := scope.Run(strings.NewReader(scriptSource(script.node)), chunkName)
		if err != nil {
			return nil, fmt.Errorf("static Lua block %d in %s failed\n%s", block, source, err.Error())
		}
		if result.Exited && result.ExitCode != 0 {
			if result.ExitCode == 1 {
				return nil, fmt.Errorf("static Lua block %d in %s exited with failure", block, source)
			}
			return nil, fmt.Errorf("static Lua block %d in %s exited with code %d", block, source, result.ExitCode)
		}
		if err := validateLuaOutput(emitter.nodes, emitter.parent, c.components); err != nil {
			return nil, fmt.Errorf("static Lua block %d in %s : %s", block, source, err.Error())
		}

		parent := script.node.Parent
		for _, node := range emitter.nodes {
			parent.InsertBefore(node, script.node)
		}
		parent.RemoveChild(script.node)
	}

	if len(dynamic.blocks) == 0 {
		return nil, nil
	}
	return dynamic, nil
}

func collectInlineLua(root *html.Node) ([]inlineLuaScript, error) {
	var scripts []inlineLuaScript
	block := 0
	var visit func(*html.Node) error
	visit = func(node *html.Node) error {
		if node.Type == html.ElementNode && node.Data == "script" {
			matched, static, err := classifyInlineLua(node)
			if err != nil {
				return fmt.Errorf("inline Lua block %d : %s", block+1, err.Error())
			}
			if matched {
				block++
				scripts = append(scripts, inlineLuaScript{node: node, static: static})
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	return scripts, nil
}

func classifyInlineLua(node *html.Node) (bool, bool, error) {
	inline := false
	typeCount := 0
	static := false
	for _, attribute := range node.Attr {
		switch attribute.Key {
		case "type":
			typeCount++
			inline = inline || attribute.Val == inlineLuaType
		case "static":
			static = true
		}
	}
	if !inline {
		return false, false, nil
	}
	if typeCount > 1 {
		return false, false, errors.New("script contains more than one type attribute")
	}
	for _, attribute := range node.Attr {
		if attribute.Key == "src" {
			return false, false, errors.New("inline Lua scripts cannot have a src attribute")
		}
	}
	return true, static, nil
}

func beneathSklairIf(node *html.Node) bool {
	for current := node; current != nil; current = current.Parent {
		for _, attribute := range current.Attr {
			if attribute.Key == "sklair-if" {
				return true
			}
		}
	}
	return false
}

func scriptSource(script *html.Node) string {
	var source strings.Builder
	for child := script.FirstChild; child != nil; child = child.NextSibling {
		source.WriteString(child.Data)
	}
	return source.String()
}
