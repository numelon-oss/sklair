package building

import (
	"fmt"
	"strconv"

	"sklair/luaSandbox"

	"golang.org/x/net/html"
)

func (c *staticLuaCompiler) runDynamic(roots []*html.Node, definition *dynamicLuaDefinition, props *boundProps, source string) error {
	if definition == nil {
		return nil
	}
	emitter := &luaEmitter{}
	scope := c.runtime.NewScope(luaSandbox.SandboxOptions{Profile: luaSandbox.InlineSandbox, FSContext: c.fsContext})
	if err := scope.SetReadOnly("props", props.luaValues()); err != nil {
		return fmt.Errorf("could not expose props to dynamic Lua in %s : %s", source, err.Error())
	}
	openSklair(scope, emitter, c)

	markers := make([]*html.Node, 0, len(definition.blocks))
	for _, root := range roots {
		collectDynamicMarkers(root, &markers)
	}
	for _, marker := range markers {
		index, err := dynamicMarkerIndex(marker)
		if err != nil || index >= len(definition.blocks) {
			return fmt.Errorf("invalid internal dynamic Lua marker in %s", source)
		}
		block := definition.blocks[index]
		emitter.parent = marker.Parent
		emitter.nodes = nil
		result, err := scope.RunPrototype(block.prototype)
		if err != nil {
			return fmt.Errorf("dynamic Lua block %d in %s failed\n%s", block.ordinal, source, err.Error())
		}
		if result.Exited && result.ExitCode != 0 {
			if result.ExitCode == 1 {
				return fmt.Errorf("dynamic Lua block %d in %s exited with failure", block.ordinal, source)
			}
			return fmt.Errorf("dynamic Lua block %d in %s exited with code %d", block.ordinal, source, result.ExitCode)
		}
		if err := validateLuaOutput(emitter.nodes, emitter.parent, c.components); err != nil {
			return fmt.Errorf("dynamic Lua block %d in %s : %s", block.ordinal, source, err.Error())
		}
		if err := props.bindEmitted(emitter.nodes); err != nil {
			return fmt.Errorf("could not bind dynamic Lua block %d in %s : %s", block.ordinal, source, err.Error())
		}
		for _, node := range emitter.nodes {
			marker.Parent.InsertBefore(node, marker)
		}
		marker.Parent.RemoveChild(marker)
	}
	return props.finish()
}

func collectDynamicMarkers(node *html.Node, markers *[]*html.Node) {
	if node.Type == html.ElementNode && node.Data == dynamicLuaMarker {
		*markers = append(*markers, node)
		return
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectDynamicMarkers(child, markers)
	}
}

func dynamicMarkerIndex(marker *html.Node) (int, error) {
	for _, attribute := range marker.Attr {
		if attribute.Key == "block" {
			return strconv.Atoi(attribute.Val)
		}
	}
	return 0, fmt.Errorf("missing block")
}
