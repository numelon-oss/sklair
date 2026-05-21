package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ComponentSource struct {
	Path     string
	IsFolder bool
}

func (s ComponentSource) Entry() string {
	if s.IsFolder {
		return filepath.Join(s.Path, "index.html")
	}

	return s.Path
}

func (s ComponentSource) AssetDir() string {
	if s.IsFolder {
		return s.Path
	}

	return ""
}

func DiscoverComponents(source string) (map[string]ComponentSource, error) {
	dir, err := os.ReadDir(source)
	if err != nil {
		return nil, err
	}

	components := make(map[string]ComponentSource)

	for _, entry := range dir {
		name := entry.Name()
		key := ""
		component := ComponentSource{}

		if entry.IsDir() {
			indexPath := filepath.Join(source, name, "index.html")
			if _, err := os.Stat(indexPath); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}

			key = strings.ToLower(name)
			component = ComponentSource{
				Path:     name,
				IsFolder: true,
			}
		} else {
			ext := filepath.Ext(name)
			key = strings.ToLower(strings.TrimSuffix(name, ext))
			component = ComponentSource{Path: name}
		}

		if previous, exists := components[key]; exists {
			return nil, fmt.Errorf("component name collision for %s between %s and %s", key, previous.Entry(), component.Entry())
		}

		components[key] = component
	}

	return components, nil
}
