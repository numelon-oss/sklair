package building

import (
	"fmt"
	"path/filepath"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/util"
	"sort"

	"golang.org/x/net/html"
)

func contributeComponent(
	component *componentInstance,
	resolver *componentResolver,
	head *html.Node,
	used map[string]struct{},
	usedFolders map[string]discovery.ComponentSource,
) {
	if _, exists := used[component.Key]; exists {
		return
	}
	used[component.Key] = struct{}{}
	for _, dependency := range component.Dependencies {
		contributeComponent(dependency, resolver, head, used, usedFolders)
	}

	htmlUtilities.AppendNodes(head, component.HeadNodes)
	if source := resolver.definitions.sources[component.Name]; source.IsFolder {
		usedFolders[component.Name] = source
	}
}

func copyComponentFolders(componentsDir string, outputDir string, components map[string]discovery.ComponentSource) error {
	names := make([]string, 0, len(components))
	for name := range components {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, tag := range names {
		component := components[tag]
		source := filepath.Join(componentsDir, component.AssetDir())
		destination := filepath.Join(outputDir, "_sklair", "components", tag)
		if err := util.CopyDir(source, destination); err != nil {
			return fmt.Errorf("could not copy component assets from %s : %s", source, err.Error())
		}
	}

	return nil
}
