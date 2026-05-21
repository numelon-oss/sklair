package building

import (
	"errors"
	"fmt"
	"path/filepath"
	"sklair/discovery"
	"sklair/util"
	"strings"

	"golang.org/x/net/html"
)

func checkNoBody(node *html.Node) error {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode && strings.TrimSpace(child.Data) == "" {
			continue
		}

		return errors.New("component bodies are not supported")
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
