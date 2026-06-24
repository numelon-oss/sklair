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

func compileDocuments(definitions *definitionSet, templates map[string]struct{}) ([]*documentState, error) {
	resolver := newComponentResolver(definitions.components, templates)
	documents := make([]*documentState, 0, len(definitions.documents))

	for _, definition := range definitions.documents {
		document, err := compileDocument(definition, resolver)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}

	return documents, nil
}

func compileDocument(definition documentDefinition, resolver *componentResolver) (*documentState, error) {
	doc := definition.root
	head := htmlUtilities.FindTag(doc, "head")
	body := htmlUtilities.FindTag(doc, "body")
	if head == nil || body == nil {
		return nil, fmt.Errorf("could not find head or body tags in %s, how does that even happen", definition.source)
	}

	state := &documentState{
		documentDefinition: definition,
		templates:          make(map[string]*componentInstance),
		componentFolders:   make(map[string]discovery.ComponentSource),
	}

	// usedComponents ensures each component and its recursive dependencies contribute
	// their <head> nodes and folder assets at most once per document
	usedComponents := make(map[string]struct{})
	explicitTemplates := make(map[string]struct{})
	count := 0
	if err := compileDocumentNodes(doc, head, definition.source, resolver, state, usedComponents, explicitTemplates, &count); err != nil {
		return nil, err
	}
	logger.Info("Found %d tags to replace in %s", count, definition.source)

	return state, nil
}

func compileDocumentNodes(
	parent *html.Node,
	head *html.Node,
	source string,
	resolver *componentResolver,
	state *documentState,
	usedComponents map[string]struct{},
	explicitTemplates map[string]struct{},
	count *int,
) error {
	for node := parent.FirstChild; node != nil; {
		next := node.NextSibling
		if node.Type != html.ElementNode {
			if err := compileDocumentNodes(node, head, source, resolver, state, usedComponents, explicitTemplates, count); err != nil {
				return err
			}
			node = next
			continue
		}

		tag := strings.ToLower(node.Data)
		if htmlUtilities.HtmlTags[tag] {
			if err := compileDocumentNodes(node, head, source, resolver, state, usedComponents, explicitTemplates, count); err != nil {
				return err
			}
			node = next
			continue
		}

		_, componentExists := resolver.definitions.sources[tag]
		if componentExists {
			(*count)++
			if htmlUtilities.HasChildren(node) {
				if _, isRuntimeTemplate := resolver.templates[tag]; isRuntimeTemplate {
					return fmt.Errorf("runtime template component %s in %s cannot receive a body yet", node.Data, source)
				}
				acceptsBody, err := resolver.acceptsBody(tag)
				if err != nil {
					return err
				}
				if !acceptsBody {
					return fmt.Errorf("invalid use of component %s in %s : component does not declare a default slot", node.Data, source)
				}
				if err := compileDocumentNodes(node, head, source, resolver, state, usedComponents, explicitTemplates, count); err != nil {
					return err
				}
			}

			resolved, err := resolver.Instantiate(tag, node.Attr, htmlUtilities.GetAllChildren(node))
			if err != nil {
				return fmt.Errorf("could not resolve component %s : %s", node.Data, err.Error())
			}
			if resolved.Dynamic {
				logger.Warning("Dynamic components are not implemented yet, skipping %s...", node.Data)
				node = next
				continue
			}

			contributeComponent(resolved, resolver, head, usedComponents, state.componentFolders)

			if _, isRuntimeTemplate := resolver.templates[tag]; isRuntimeTemplate {
				if _, registered := explicitTemplates[tag]; registered {
					return fmt.Errorf("runtime template component %s is registered more than once in %s", node.Data, source)
				}
				explicitTemplates[tag] = struct{}{}
				if err := addTemplate(state.templates, resolved); err != nil {
					return fmt.Errorf("could not register runtime template %s in %s : %s", node.Data, source, err.Error())
				}
			} else {
				htmlUtilities.InsertNodesBefore(node, resolved.BodyNodes)
			}
			if err := mergeTemplates(state.templates, resolved.RuntimeTemplates); err != nil {
				return fmt.Errorf("could not register runtime template dependency in %s : %s", source, err.Error())
			}

			node.Parent.RemoveChild(node)
			node = next
			continue
		}

		switch tag {
		case "lua":
			(*count)++
			// TODO: prints from lua will be appended to a buffer
			// then this buffer will be parsed by html
			// then this will be inserted into document
			// TODO: or should we actually instead expose a library eg `sklair` and we can do `sklair.put()`? thats probably cleaner
			// and also easier to implement
			logger.Warning("Lua components for regular input files are not implemented yet, skipping...")

		case "opengraph":
			(*count)++
			for _, child := range snippets.OpenGraph(node) {
				head.AppendChild(child)
			}
			node.Parent.RemoveChild(node)

		default:
			logger.Warning("Non-standard tag found in HTML and no component present : %s; assuming Autonomous Custom Element", tag)
			if err := compileDocumentNodes(node, head, source, resolver, state, usedComponents, explicitTemplates, count); err != nil {
				return err
			}
		}

		node = next
	}

	return nil
}
