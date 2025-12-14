package discovery

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

type DocumentLists struct {
	HtmlFiles   []string
	StaticFiles []string
}

var defaultExcludes = []string{
	"**/components/**",
	"**/.git/**",
	"**/.vscode/**",
	"**/.idea/**",
	"**/.env",
	"**/node_modules/**",
	"**/build/**",
	"**/sklair.json",
}

func splitPatterns(patterns []string) (excludes, includes []string) {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			includes = append(includes, pattern[1:])
		} else {
			excludes = append(excludes, pattern)
		}
	}

	return excludes, includes
}

func isExcluded(rel string, excludes []string, includes []string) bool {
	rel = filepath.ToSlash(rel)

	for _, pattern := range excludes {
		if matched, _ := doublestar.PathMatch(pattern, rel); matched {
			// check if overridden by an include pattern
			for _, include := range includes {
				if undo, _ := doublestar.PathMatch(include, rel); undo {
					return false
				}
			}

			return true
		}
	}

	return false
}

func DocumentDiscovery(root string, excludes []string) (*DocumentLists, error) {
	lists := &DocumentLists{}

	excludes = append(defaultExcludes, excludes...)
	excludePatterns, includePatterns := splitPatterns(excludes)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		if relPath == "." {
			return nil // NEVER exclude root!!
		}

		// doublestar excludes
		if isExcluded(relPath, excludePatterns, includePatterns) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// will be walked by filepath.Walk later anyway
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(strings.ToLower(info.Name()))
		// TODO: perhaps allow this file ext to be customisable?
		if ext == ".html" || ext == ".shtml" {
			lists.HtmlFiles = append(lists.HtmlFiles, path)
		} else {
			lists.StaticFiles = append(lists.StaticFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return lists, nil
}
