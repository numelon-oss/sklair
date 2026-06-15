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

func (r *componentResolver) Instantiate(name string, attributes []html.Attribute) (*componentInstance, error) {
	return r.instantiate(strings.ToLower(name), attributes, nil)
}

func (r *componentResolver) instantiate(name string, attributes []html.Attribute, stack []string) (*componentInstance, error) {
	for _, ancestor := range stack {
		if ancestor == name {
			return nil, circularDepErr(stack, name)
		}
	}

	definition, source, err := r.definitions.prepare(name)
	if err != nil {
		return nil, err
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

	stack = append(stack, name)
	var invocations []*html.Node
	for node := range body.Descendants() {
		if node.Type != html.ElementNode {
			continue
		}

		tag := strings.ToLower(node.Data)
		if htmlUtilities.HtmlTags[tag] || tag == "lua" || tag == "opengraph" {
			continue
		}
		if _, exists := r.definitions.sources[tag]; !exists {
			logger.Warning("Non-standard tag found in component %s and no component present : %s; assuming Autonomous Custom Element", name, tag)
			continue
		}
		invocations = append(invocations, node)
	}

	explicitTemplates := make(map[string]struct{})
	for _, invocation := range invocations {
		if htmlUtilities.HasChildren(invocation) {
			return nil, fmt.Errorf("invalid use of component %s inside %s : component bodies are not supported", invocation.Data, name)
		}

		dependencyName := strings.ToLower(invocation.Data)
		dependency, err := r.instantiate(dependencyName, invocation.Attr, stack)
		if err != nil {
			return nil, err
		}
		instance.Dependencies = append(instance.Dependencies, dependency)

		if _, isRuntimeTemplate := r.templates[dependencyName]; isRuntimeTemplate {
			if _, exists := explicitTemplates[dependencyName]; exists {
				return nil, fmt.Errorf("runtime template component %s is registered more than once inside %s", invocation.Data, name)
			}
			explicitTemplates[dependencyName] = struct{}{}
			if err := addTemplate(instance.RuntimeTemplates, dependency); err != nil {
				return nil, err
			}
		} else {
			htmlUtilities.InsertNodesBefore(invocation, dependency.BodyNodes)
		}
		if err := mergeTemplates(instance.RuntimeTemplates, dependency.RuntimeTemplates); err != nil {
			return nil, err
		}

		instance.Dynamic = instance.Dynamic || dependency.Dynamic
		invocation.Parent.RemoveChild(invocation)
	}

	instance.BodyNodes = htmlUtilities.GetAllChildren(body)
	if _, isRuntimeTemplate := r.templates[name]; isRuntimeTemplate && instance.Dynamic {
		return nil, fmt.Errorf("runtime template component %s is dynamic, but dynamic components are not implemented yet", name)
	}
	return instance, nil
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
