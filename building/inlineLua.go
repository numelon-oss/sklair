package building

import (
	"errors"
	"fmt"
	"sklair/discovery"
	"sklair/luaSandbox"
	"strings"

	"golang.org/x/net/html"
)

const inlineLuaType = "application/x-sklair-lua"

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

func (c *staticLuaCompiler) prepare(root *html.Node, source string) error {
	scripts, err := collectInlineLua(root)
	if err != nil {
		return err
	}
	if len(scripts) == 0 {
		return nil
	}

	for i, script := range scripts {
		if !script.static {
			return fmt.Errorf("dynamic inline Lua block %d is not implemented until Stage 5", i+1)
		}
		if beneathSklairIf(script.node) {
			return fmt.Errorf("static Lua block %d cannot be nested beneath sklair-if", i+1)
		}
	}

	emitter := &luaEmitter{}
	scope := c.runtime.NewScope(luaSandbox.SandboxOptions{
		Profile:   luaSandbox.InlineSandbox,
		FSContext: c.fsContext,
	})
	openSklair(scope, emitter, c)

	for i, script := range scripts {
		emitter.parent = script.node.Parent
		emitter.nodes = nil
		block := i + 1
		chunkName := fmt.Sprintf("%s:static-lua-%d", source, block)
		result, err := scope.Run(strings.NewReader(scriptSource(script.node)), chunkName)
		if err != nil {
			return fmt.Errorf("static Lua block %d in %s failed\n%s", block, source, err.Error())
		}
		if result.Exited && result.ExitCode != 0 {
			if result.ExitCode == 1 {
				return fmt.Errorf("static Lua block %d in %s exited with failure", block, source)
			}
			return fmt.Errorf("static Lua block %d in %s exited with code %d", block, source, result.ExitCode)
		}
		if err := validateLuaOutput(emitter.nodes, emitter.parent, c.components); err != nil {
			return fmt.Errorf("static Lua block %d in %s : %s", block, source, err.Error())
		}

		parent := script.node.Parent
		for _, node := range emitter.nodes {
			parent.InsertBefore(node, script.node)
		}
		parent.RemoveChild(script.node)
	}

	return nil
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
