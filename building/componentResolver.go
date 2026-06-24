package building

import (
	"fmt"
	"sklair/building/resources"
	"sklair/htmlUtilities"
	"sort"
	"strings"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
)

type componentInstance struct {
	Name             string
	Key              string
	Props            string
	Path             []string
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
	return r.instantiate(strings.ToLower(name), attributes, body, nil)
}

func (r *componentResolver) acceptsBody(name string) (bool, error) {
	definition, _, err := r.definitions.prepare(strings.ToLower(name))
	if err != nil {
		return false, err
	}
	return findSlot(definition.body) != nil, nil
}

func (r *componentResolver) instantiate(name string, attributes []html.Attribute, projectedBody []*html.Node, stack []string) (*componentInstance, error) {
	for _, ancestor := range stack {
		if ancestor == name {
			return nil, circularDepErr(stack, name)
		}
	}

	definition, source, err := r.definitions.prepare(name)
	if err != nil {
		return nil, err
	}
	if len(projectedBody) > 0 && findSlot(definition.body) == nil {
		return nil, fmt.Errorf("component %s does not declare a default slot", name)
	}
	if len(projectedBody) > 0 {
		if _, isRuntimeTemplate := r.templates[name]; isRuntimeTemplate {
			return nil, fmt.Errorf("runtime template component %s cannot receive a body yet", name)
		}
	}

	head := &html.Node{Type: html.DocumentNode}
	for _, node := range definition.head {
		head.AppendChild(htmlUtilities.Clone(node))
	}
	body := &html.Node{Type: html.DocumentNode}
	for _, node := range definition.body {
		body.AppendChild(htmlUtilities.Clone(node))
	}

	allNodes := append(htmlUtilities.GetAllChildren(head), htmlUtilities.GetAllChildren(body)...)
	props, err := bind(allNodes, attributes)
	if err != nil {
		return nil, fmt.Errorf("could not bind component %s : %s", name, err.Error())
	}

	instance := &componentInstance{
		Name:             name,
		Key:              name + "\x00" + props.signature,
		Props:            props.description,
		Path:             append(append([]string{}, stack...), name),
		HeadNodes:        htmlUtilities.GetAllChildren(head),
		RuntimeTemplates: make(map[string]*componentInstance),
		Dynamic:          definition.dynamic,
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

	slot := findSlot(htmlUtilities.GetAllChildren(body))
	if len(projectedBody) > 0 && slot == nil {
		return nil, fmt.Errorf("component %s default slot is not available for these props", name)
	}
	if slot != nil {
		htmlUtilities.InsertNodesBefore(slot, projectedBody)
		slot.Parent.RemoveChild(slot)
	}

	stack = append(stack, name)
	explicitTemplates := make(map[string]struct{})
	if err := r.resolveNodes(body, name, stack, instance, explicitTemplates); err != nil {
		return nil, err
	}

	instance.BodyNodes = htmlUtilities.GetAllChildren(body)
	if _, isRuntimeTemplate := r.templates[name]; isRuntimeTemplate && instance.Dynamic {
		return nil, fmt.Errorf("runtime template component %s is dynamic, but dynamic components are not implemented yet", name)
	}
	return instance, nil
}

func (r *componentResolver) resolveNodes(parent *html.Node, owner string, stack []string, instance *componentInstance, explicitTemplates map[string]struct{}) error {
	for node := parent.FirstChild; node != nil; {
		next := node.NextSibling
		if node.Type != html.ElementNode {
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates); err != nil {
				return err
			}
			node = next
			continue
		}

		dependencyName := strings.ToLower(node.Data)
		if htmlUtilities.HtmlTags[dependencyName] {
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates); err != nil {
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
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates); err != nil {
				return err
			}
			node = next
			continue
		}

		if htmlUtilities.HasChildren(node) {
			if _, isRuntimeTemplate := r.templates[dependencyName]; isRuntimeTemplate {
				return fmt.Errorf("runtime template component %s inside %s cannot receive a body yet", node.Data, owner)
			}
			acceptsBody, err := r.acceptsBody(dependencyName)
			if err != nil {
				return err
			}
			if !acceptsBody {
				return fmt.Errorf("invalid use of component %s inside %s : component does not declare a default slot", node.Data, owner)
			}
			if err := r.resolveNodes(node, owner, stack, instance, explicitTemplates); err != nil {
				return err
			}
		}

		dependency, err := r.instantiate(dependencyName, node.Attr, htmlUtilities.GetAllChildren(node), stack)
		if err != nil {
			return err
		}
		instance.Dependencies = append(instance.Dependencies, dependency)

		if _, isRuntimeTemplate := r.templates[dependencyName]; isRuntimeTemplate {
			if _, exists := explicitTemplates[dependencyName]; exists {
				return fmt.Errorf("runtime template component %s is registered more than once inside %s", node.Data, owner)
			}
			explicitTemplates[dependencyName] = struct{}{}
			if err := addTemplate(instance.RuntimeTemplates, dependency); err != nil {
				return err
			}
		} else {
			htmlUtilities.InsertNodesBefore(node, dependency.BodyNodes)
		}
		if err := mergeTemplates(instance.RuntimeTemplates, dependency.RuntimeTemplates); err != nil {
			return err
		}

		instance.Dynamic = instance.Dynamic || dependency.Dynamic
		node.Parent.RemoveChild(node)
		node = next
	}

	return nil
}

func addTemplate(templates map[string]*componentInstance, template *componentInstance) error {
	if previous, exists := templates[template.Name]; exists {
		if previous.Key != template.Key {
			return fmt.Errorf(
				"runtime template %s is required with conflicting compile-time props %s via %s and %s via %s",
				template.Name,
				previous.Props,
				strings.Join(previous.Path, " -> "),
				template.Props,
				strings.Join(template.Path, " -> "),
			)
		}
		return nil
	}

	templates[template.Name] = template
	return nil
}

func mergeTemplates(destination map[string]*componentInstance, source map[string]*componentInstance) error {
	names := make([]string, 0, len(source))
	for name := range source {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		template := source[name]
		if err := addTemplate(destination, template); err != nil {
			return err
		}
	}
	return nil
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
