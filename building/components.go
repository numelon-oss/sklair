package building

import (
	"fmt"
	"path/filepath"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/util"

	"golang.org/x/net/html"
)

func contributeComponent(
	name string,
	resolver *componentResolver,
	head *html.Node,
	used map[string]struct{},
	usedFolders map[string]discovery.ComponentSource,
) error {
	if _, exists := used[name]; exists {
		return nil
	}
	used[name] = struct{}{}

	component, err := resolver.Resolve(name)
	if err != nil {
		return err
	}
	for _, dependency := range component.Dependencies {
		if err := contributeComponent(dependency, resolver, head, used, usedFolders); err != nil {
			return err
		}
	}

	htmlUtilities.AppendNodes(head, component.HeadNodes)
	if source := resolver.sources[name]; source.IsFolder {
		usedFolders[name] = source
	}
	return nil
}

func copyComponentFolders(componentsDir string, outputDir string, components map[string]discovery.ComponentSource) error {
	for tag, component := range components {
		source := filepath.Join(componentsDir, component.AssetDir())
		destination := filepath.Join(outputDir, "_sklair", "components", tag)
		if err := util.CopyDir(source, destination); err != nil {
			return fmt.Errorf("could not copy component assets from %s : %s", source, err.Error())
		}
	}

	return nil
}
