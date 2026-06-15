package building

import (
	"fmt"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/snippets"
	"strings"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
)

type documentState struct {
	documentDefinition
	templates        map[string]*componentInstance
	componentFolders map[string]discovery.ComponentSource
}

func compileDocument(definition documentDefinition, resolver *componentResolver) (*documentState, error) {
	doc := definition.root
	head := htmlUtilities.FindTag(doc, "head")
	body := htmlUtilities.FindTag(doc, "body")
	if head == nil || body == nil {
		return nil, fmt.Errorf("could not find head or body tags in %s, how does that even happen", definition.source)
	}

	var invocations []*html.Node
	for node := range doc.Descendants() {
		if node.Type != html.ElementNode {
			continue
		}

		tag := strings.ToLower(node.Data)
		if htmlUtilities.HtmlTags[tag] {
			continue
		}
		if tag == "lua" || tag == "opengraph" {
			invocations = append(invocations, node)
			continue
		}
		if _, exists := resolver.definitions.sources[tag]; !exists {
			logger.Warning("Non-standard tag found in HTML and no component present : %s; assuming Autonomous Custom Element", tag)
			continue
		}

		invocations = append(invocations, node)
	}

	logger.Info("Found %d tags to replace in %s", len(invocations), definition.source)

	state := &documentState{
		documentDefinition: definition,
		templates:          make(map[string]*componentInstance),
		componentFolders:   make(map[string]discovery.ComponentSource),
	}

	// usedComponents ensures each component and its recursive dependencies contribute
	// their <head> nodes and folder assets at most once per document
	usedComponents := make(map[string]struct{})
	explicitTemplates := make(map[string]struct{})
	for _, invocation := range invocations {
		tag := strings.ToLower(invocation.Data)

		parent := invocation.Parent
		if parent == nil {
			return nil, fmt.Errorf("somehow the parent does not exist for %s. (memory corruption???)", invocation.Data)
		}

		if _, componentExists := resolver.definitions.sources[tag]; componentExists {
			if htmlUtilities.HasChildren(invocation) {
				return nil, fmt.Errorf("invalid use of component %s in %s : component bodies are not supported", invocation.Data, definition.source)
			}

			resolved, err := resolver.Instantiate(tag, invocation.Attr)
			if err != nil {
				return nil, fmt.Errorf("could not resolve component %s : %s", invocation.Data, err.Error())
			}
			if resolved.Dynamic {
				logger.Warning("Dynamic components are not implemented yet, skipping %s...", invocation.Data)
				continue
			}

			contributeComponent(resolved, resolver, head, usedComponents, state.componentFolders)

			if _, isRuntimeTemplate := resolver.templates[tag]; isRuntimeTemplate {
				if _, registered := explicitTemplates[tag]; registered {
					return nil, fmt.Errorf("runtime template component %s is registered more than once in %s", invocation.Data, definition.source)
				}
				explicitTemplates[tag] = struct{}{}
				if err := addTemplate(state.templates, resolved); err != nil {
					return nil, fmt.Errorf("could not register runtime template %s in %s : %s", invocation.Data, definition.source, err.Error())
				}
			} else {
				htmlUtilities.InsertNodesBefore(invocation, resolved.BodyNodes)
			}
			if err := mergeTemplates(state.templates, resolved.RuntimeTemplates); err != nil {
				return nil, fmt.Errorf("could not register runtime template dependency in %s : %s", definition.source, err.Error())
			}

			parent.RemoveChild(invocation)
			continue
		}

		switch tag {
		case "lua":
			// TODO: prints from lua will be appended to a buffer
			// then this buffer will be parsed by html
			// then this will be inserted into document
			// TODO: or should we actually instead expose a library eg `sklair` and we can do `sklair.put()`? thats probably cleaner
			// and also easier to implement
			logger.Warning("Lua components for regular input files are not implemented yet, skipping...")

		case "opengraph":
			for _, child := range snippets.OpenGraph(invocation) {
				head.AppendChild(child)
			}
			parent.RemoveChild(invocation)
		}
	}

	return state, nil
}
