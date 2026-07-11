package building

import (
	"fmt"
	"sklair/building/resources"
	"sklair/htmlUtilities"
	"slices"
	"strings"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type componentInstance struct {
	Name             string
	Key              string
	HeadNodes        []*html.Node
	BodyNodes        []*html.Node
	Dependencies     []*componentInstance
	RuntimeTemplates map[string]*componentInstance
	Dynamic          bool
}

type componentResolver struct {
	definitions *componentDefinitions
	templates   map[string]struct{}
}

func newComponentResolver(definitions *componentDefinitions, templates map[string]struct{}) *componentResolver {
	return &componentResolver{
		definitions: definitions,
		templates:   templates,
	}
}

func (r *componentResolver) Instantiate(name string, attributes []html.Attribute, body []*html.Node) (*componentInstance, error) {
	return r.instantiate(strings.ToLower(name), attributes, body, nil, false)
}

func (r *componentResolver) acceptsBody(name string) (bool, error) {
	definition, _, err := r.definitions.prepare(strings.ToLower(name))
	if err != nil {
		return false, err
	}
	return findSlot(definition.body) != nil, nil
}

func (r *componentResolver) instantiate(name string, attributes []html.Attribute, projectedBody []*html.Node, stack []string, runtimeTree bool) (*componentInstance, error) {
	if slices.Contains(stack, name) {
		return nil, circularDepErr(stack, name)
	}

	definition, source, err := r.definitions.prepare(name)
	if err != nil {
		return nil, err
	}
	_, isRuntimeTemplate := r.templates[name]
	if runtimeTree && definition.lua != nil {
		return nil, fmt.Errorf("runtime template dependency %s contains compile-time dynamic Lua", name)
	}
	runtimeTree = runtimeTree || isRuntimeTemplate
	if err := validateComponentMode(definition, isRuntimeTemplate); err != nil {
		return nil, fmt.Errorf("invalid component %s : %s", name, err.Error())
	}
	if isRuntimeTemplate && len(projectedBody) > 0 {
		return nil, fmt.Errorf("runtime template component %s registration must be empty", name)
	}
	if isRuntimeTemplate && len(attributes) > 0 {
		return nil, fmt.Errorf("runtime template component %s registration cannot have attributes", name)
	}
	if !isRuntimeTemplate && len(projectedBody) > 0 && findSlot(definition.body) == nil {
		return nil, fmt.Errorf("component %s does not declare a default slot", name)
	}

	head := &html.Node{Type: html.ElementNode, DataAtom: atom.Head, Data: "head"}
	for _, node := range definition.head {
		head.AppendChild(htmlUtilities.Clone(node))
	}
	body := &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"}
	for _, node := range definition.body {
		body.AppendChild(htmlUtilities.Clone(node))
	}

	key := name
	var props *componentProps
	if !isRuntimeTemplate {
		allNodes := append(htmlUtilities.GetAllChildren(head), htmlUtilities.GetAllChildren(body)...)
		props, err = bind(allNodes, attributes, definition.lua)
		if err != nil {
			return nil, fmt.Errorf("could not bind component %s : %s", name, err.Error())
		}
		key += "\x00" + props.signature
	}

	instance := &componentInstance{
		Name:             name,
		Key:              key,
		HeadNodes:        htmlUtilities.GetAllChildren(head),
		RuntimeTemplates: make(map[string]*componentInstance),
		Dynamic:          definition.lua != nil,
	}

	if !isRuntimeTemplate {
		slot := findSlot(htmlUtilities.GetAllChildren(body))
		if len(projectedBody) > 0 && slot == nil {
			return nil, fmt.Errorf("component %s default slot is not available for these props", name)
		}
		if slot != nil {
			htmlUtilities.InsertNodesBefore(slot, projectedBody)
			slot.Parent.RemoveChild(slot)
		}
	}
	if !isRuntimeTemplate {
		roots := append(htmlUtilities.GetAllChildren(head), htmlUtilities.GetAllChildren(body)...)
		if err := r.definitions.static.runDynamic(roots, definition.lua, props, source.Entry()); err != nil {
			return nil, err
		}
		if err := props.normalise(); err != nil {
			return nil, fmt.Errorf("could not normalise component %s props : %s", name, err.Error())
		}
		instance.Key = name + "\x00" + props.signature
		instance.HeadNodes = htmlUtilities.GetAllChildren(head)
	}

	if source.IsFolder {
		outputDir := "/_sklair/components/" + name
		if err := resources.RewriteURLs(instance.HeadNodes, outputDir); err != nil {
			return nil, fmt.Errorf("could not rewrite component %s head resource URLs : %s", source.Entry(), err.Error())
		}
		if err := resources.RewriteURLs(htmlUtilities.GetAllChildren(body), outputDir); err != nil {
			return nil, fmt.Errorf("could not rewrite component %s body resource URLs : %s", source.Entry(), err.Error())
		}
	}

	stack = append(stack, name)
	explicitTemplates := make(map[string]struct{})
	if err := r.resolveNodes(body, name, stack, instance, explicitTemplates, runtimeTree); err != nil {
		return nil, err
	}
	if isRuntimeTemplate {
		if err := makeRuntime(htmlUtilities.GetAllChildren(body)); err != nil {
			return nil, fmt.Errorf("could not prepare runtime bindings in component %s : %s", name, err.Error())
		}
	}

	instance.BodyNodes = htmlUtilities.GetAllChildren(body)
	if isRuntimeTemplate && instance.Dynamic {
		return nil, fmt.Errorf("runtime template component %s contains compile-time dynamic Lua", name)
	}
	return instance, nil
}

func (r *componentResolver) resolveNodes(parent *html.Node, owner string, stack []string, instance *componentInstance, explicitTemplates map[string]struct{}, runtimeTree bool) error {
	for node := parent.FirstChild; node != nil; {
		next := node.NextSibling
		if node.Type != html.ElementNode {
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates, runtimeTree); err != nil {
				return err
			}
			node = next
			continue
		}

		dependencyName := strings.ToLower(node.Data)
		if htmlUtilities.HtmlTags[dependencyName] {
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates, runtimeTree); err != nil {
				return err
			}
			node = next
			continue
		}

		_, componentExists := r.definitions.sources[dependencyName]
		if !componentExists {
			if dependencyName != "lua" && dependencyName != "opengraph" {
				logger.Warning("Non-standard tag found in component %s and no component present : %s; assuming Autonomous Custom Element", owner, dependencyName)
			}
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates, runtimeTree); err != nil {
				return err
			}
			node = next
			continue
		}

		if htmlUtilities.HasChildren(node) {
			if _, isRuntimeTemplate := r.templates[dependencyName]; isRuntimeTemplate {
				return fmt.Errorf("runtime template component %s registration inside %s must be empty", node.Data, owner)
			}
			acceptsBody, err := r.acceptsBody(dependencyName)
			if err != nil {
				return err
			}
			if !acceptsBody {
				return fmt.Errorf("invalid use of component %s inside %s : component does not declare a default slot", node.Data, owner)
			}
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates, runtimeTree); err != nil {
				return err
			}
		}

		dependency, err := r.instantiate(dependencyName, node.Attr, htmlUtilities.GetAllChildren(node), stack, runtimeTree)
		if err != nil {
			return err
		}
		instance.Dependencies = append(instance.Dependencies, dependency)

		if _, isRuntimeTemplate := r.templates[dependencyName]; isRuntimeTemplate {
			if _, exists := explicitTemplates[dependencyName]; exists {
				return fmt.Errorf("runtime template component %s is registered more than once inside %s", node.Data, owner)
			}
			explicitTemplates[dependencyName] = struct{}{}
			addTemplate(instance.RuntimeTemplates, dependency)
		} else {
			htmlUtilities.InsertNodesBefore(node, dependency.BodyNodes)
		}
		mergeTemplates(instance.RuntimeTemplates, dependency.RuntimeTemplates)

		instance.Dynamic = instance.Dynamic || dependency.Dynamic
		node.Parent.RemoveChild(node)
		node = next
	}

	return nil
}

func addTemplate(templates map[string]*componentInstance, template *componentInstance) {
	if _, exists := templates[template.Name]; exists {
		return
	}

	templates[template.Name] = template
}

func mergeTemplates(destination map[string]*componentInstance, source map[string]*componentInstance) {
	for _, template := range source {
		addTemplate(destination, template)
	}
}

func circularDepErr(stack []string, name string) error {
	start := 0
	for i, component := range stack {
		if component == name {
			start = i
			break
		}
	}

	cycle := append([]string{}, stack[start:]...)
	cycle = append(cycle, name)
	return fmt.Errorf("circular component dependency : %s", strings.Join(cycle, " -> "))
}
