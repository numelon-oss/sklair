package discovery

import (
	"os"
	"path/filepath"
	"strings"
)

type DocumentLists struct {
	HtmlFiles   []string
	StaticFiles []string
}

var skipDirs = map[string]struct{}{
	"components": {},
	".git":       {},
	".vscode":    {},
	".idea":      {},
	".env":       {}, // these are usually files for most people,
	// but at Numelon these are folders and therefore must be excluded (for safety)
	"node_modules": {},
	"build":        {},
}

var skipFiles = map[string]struct{}{
	"sklair.json": {},
	".env":        {},
}

// TODO: ensure that sklair.json is respected here
// TODO: also ensure that sklair.json input, output, components and build paths are used here instead of predefined values (or in addition to predefined values)
func DocumentDiscovery(root string) (*DocumentLists, error) {
	lists := &DocumentLists{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if _, ok := skipDirs[filepath.Base(path)]; ok && info.IsDir() {
			return filepath.SkipDir
		}

		// skip directories since they will be walked by filepath.Walk later anyway
		if info.IsDir() {
			return nil
		}

		fileName := strings.ToLower(info.Name())
		ext := filepath.Ext(fileName)
		// TODO: perhaps allow this file ext to be customisable?
		if ext == ".html" || ext == ".shtml" {
			lists.HtmlFiles = append(lists.HtmlFiles, path)
		} else if _, ok := skipFiles[fileName]; !ok {
			lists.StaticFiles = append(lists.StaticFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}
