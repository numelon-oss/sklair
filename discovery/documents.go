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
	"sklair.json",
	".sklair/",
	".git/",
	".vscode/",
	".idea/",
	".env*",
	"node_modules/",
	".DS_*",
	"._*",
}

func normaliseExcludes(patterns []string) []string {
	var out []string

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		negated := strings.HasPrefix(p, "!")
		if negated {
			p = p[1:]
		}

		p = filepath.ToSlash(p)

		rootAnchored := strings.HasPrefix(p, "/")
		if rootAnchored {
			p = p[1:]
		}

		// directory
		if strings.HasSuffix(p, "/") {
			p = strings.TrimSuffix(p, "/")
			if rootAnchored {
				p = p + "/**"
			} else {
				p = "**/" + p + "/**"
			}
		} else {
			// file or glob
			if !rootAnchored {
				p = "**/" + p
			}
		}

		if negated {
			p = "!" + p
		}

		out = append(out, p)
	}

	return out
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

func DiscoverDocuments(root string, excludes []string, excludeCompile []string) (*DocumentLists, error) {
	lists := &DocumentLists{}

	excludes = append(defaultExcludes, excludes...)
	excludes = normaliseExcludes(excludes)
	//fmt.Println(excludes)
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
		if (ext == ".htm" || ext == ".html" || ext == ".shtml" || ext == ".xhtml") && !isExcluded(relPath, excludeCompile, includePatterns) {
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
