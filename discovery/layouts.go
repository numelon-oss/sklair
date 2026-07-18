package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverLayouts(source string) (map[string]string, error) {
	entries, err := os.ReadDir(source)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}

	layouts := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		extension := strings.ToLower(filepath.Ext(name))
		switch extension {
		case ".htm", ".html", ".shtml", ".xhtml":
		default:
			continue
		}

		key := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))
		if previous, exists := layouts[key]; exists {
			return nil, fmt.Errorf("layout name collision for %s between %s and %s", key, previous, name)
		}
		layouts[key] = name
	}

	return layouts, nil
}
