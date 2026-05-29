package building

import (
	"fmt"
	"sklair/building/resources"
	"sklair/caching"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"strings"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
)

type componentResolveState uint8

const (
	componentUnseen componentResolveState = iota
	componentVisiting
	componentResolved
)

type resolvedComponent struct {
	HeadNodes        []*html.Node
	BodyNodes        []*html.Node
	Dependencies     []string
	RuntimeTemplates map[string]struct{}
	Dynamic          bool
}

type componentResolver struct {
	componentsDir string
	sources       map[string]discovery.ComponentSource
	templates     map[string]struct{}

	rawCache caching.ComponentCache
	resolved map[string]*resolvedComponent
	states   map[string]componentResolveState
	stack    []string
}

func newComponentResolver(componentsDir string, sources map[string]discovery.ComponentSource, templates map[string]struct{}) *componentResolver {
	return &componentResolver{
		componentsDir: componentsDir,
		sources:       sources,
		templates:     templates,
		rawCache: caching.ComponentCache{
			Static:  make(map[string]*caching.Component),
			Dynamic: make(map[string]*caching.Component),
		},
		resolved: make(map[string]*resolvedComponent),
		states:   make(map[string]componentResolveState),
	}
}

func (r *componentResolver) Resolve(name string) (*resolvedComponent, error) {
	name = strings.ToLower(name)

	switch r.states[name] {
	case componentResolved:
		return r.resolved[name], nil
	case componentVisiting:
		return nil, r.circularDepErr(name)
	}

	source, exists := r.sources[name]
	if !exists {
		return nil, fmt.Errorf("component %q does not exist", name)
	}

	r.states[name] = componentVisiting
	r.stack = append(r.stack, name)
	succeeded := false
	defer func() {
		r.stack = r.stack[:len(r.stack)-1]
		if !succeeded {
			r.states[name] = componentUnseen
		}
	}()

	raw, err := r.loadRaw(name, source)
	if err != nil {
		return nil, err
	}

	resolved := &resolvedComponent{
		HeadNodes:        make([]*html.Node, 0, len(raw.HeadNodes)),
		RuntimeTemplates: make(map[string]struct{}),
		Dynamic:          raw.Dynamic,
	}
	for _, node := range raw.HeadNodes {
		resolved.HeadNodes = append(resolved.HeadNodes, htmlUtilities.Clone(node))
	}

	body := &html.Node{Type: html.DocumentNode}
	for _, node := range raw.BodyNodes {
		body.AppendChild(htmlUtilities.Clone(node))
	}

	if source.IsFolder {
		outputDir := "/_sklair/components/" + name
		if err := resources.RewriteURLs(resolved.HeadNodes, outputDir); err != nil {
			return nil, fmt.Errorf("could not rewrite component %s head resource URLs : %s", source.Entry(), err.Error())
		}
		if err := resources.RewriteURLs(htmlUtilities.GetAllChildren(body), outputDir); err != nil {
			return nil, fmt.Errorf("could not rewrite component %s body resource URLs : %s", source.Entry(), err.Error())
		}
	}

	var invocations []*html.Node
	for node := range body.Descendants() {
		if node.Type != html.ElementNode {
			continue
		}

		tag := strings.ToLower(node.Data)
		if htmlUtilities.HtmlTags[tag] || tag == "lua" || tag == "opengraph" {
			continue
		}

		if _, exists := r.sources[tag]; !exists {
			logger.Warning("Non-standard tag found in component %s and no component present : %s; assuming Autonomous Custom Element", name, tag)
			continue
		}

		invocations = append(invocations, node)
	}

	dependencySeen := make(map[string]struct{})
	explicitTemplateSeen := make(map[string]struct{})
	for _, invocation := range invocations {
		if htmlUtilities.HasChildren(invocation) {
			return nil, fmt.Errorf("invalid use of component %s inside %s : component bodies are not supported", invocation.Data, name)
		}

		dependencyName := strings.ToLower(invocation.Data)
		dependency, err := r.Resolve(dependencyName)
		if err != nil {
			return nil, err
		}

		_, isRuntimeTemplate := r.templates[dependencyName]
		if _, seen := dependencySeen[dependencyName]; !seen {
			resolved.Dependencies = append(resolved.Dependencies, dependencyName)
			dependencySeen[dependencyName] = struct{}{}
		}

		if isRuntimeTemplate {
			if _, exists := explicitTemplateSeen[dependencyName]; exists {
				return nil, fmt.Errorf("runtime template component %s is registered more than once inside %s", invocation.Data, name)
			}
			explicitTemplateSeen[dependencyName] = struct{}{}
			resolved.RuntimeTemplates[dependencyName] = struct{}{}
		} else {
			htmlUtilities.InsertNodesBefore(invocation, dependency.BodyNodes)
		}
		for template := range dependency.RuntimeTemplates {
			resolved.RuntimeTemplates[template] = struct{}{}
		}

		resolved.Dynamic = resolved.Dynamic || dependency.Dynamic
		invocation.Parent.RemoveChild(invocation)
	}

	resolved.BodyNodes = htmlUtilities.GetAllChildren(body)
	if _, isRuntimeTemplate := r.templates[name]; isRuntimeTemplate && resolved.Dynamic {
		return nil, fmt.Errorf("runtime template component %s is dynamic, but dynamic components are not implemented yet", name)
	}

	r.resolved[name] = resolved
	r.states[name] = componentResolved
	succeeded = true
	return resolved, nil
}

func (r *componentResolver) loadRaw(name string, source discovery.ComponentSource) (*caching.Component, error) {
	if component, exists := r.rawCache.Static[name]; exists {
		return component, nil
	}
	if component, exists := r.rawCache.Dynamic[name]; exists {
		return component, nil
	}

	component, err := caching.MakeCache(r.componentsDir, source.Entry())
	if err != nil {
		return nil, fmt.Errorf("could not cache component %s : %s", source.Entry(), err.Error())
	}

	if component.Dynamic {
		r.rawCache.Dynamic[name] = component
	} else {
		r.rawCache.Static[name] = component
	}
	return component, nil
}

func (r *componentResolver) circularDepErr(name string) error {
	start := 0
	for i, component := range r.stack {
		if component == name {
			start = i
			break
		}
	}

	cycle := append([]string{}, r.stack[start:]...)
	cycle = append(cycle, name)
	return fmt.Errorf("circular component dependency : %s", strings.Join(cycle, " -> "))
}
