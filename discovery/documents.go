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

		// detect bare globs
		// i.e. just "test", doesn't contain or end with /, and no glob characters
		isBare := !strings.ContainsAny(p, "/*?[]")

		// exploit directory, "test/"
		if strings.HasSuffix(p, "/") {
			p = strings.TrimSuffix(p, "/")
			if rootAnchored {
				//p = p + "/**"
				out = append(out, p+"/**")
			} else {
				//p = "**/" + p + "/**"
				out = append(out, "**/"+p+"/**")
			}
		} else if isBare {
			// implicit directory, i.e. just "test"
			if rootAnchored {
				// append both, because when it is bare, we don't know whether
				// it's a file or a directory
				out = append(out, p)
				out = append(out, p+"/**")
			} else {
				out = append(out, "**/"+p)
				out = append(out, "**/"+p+"/**")
			}
		} else {
			// file or glob
			if rootAnchored {
				out = append(out, p)
			} else {
				out = append(out, "**/"+p)
			}
		}

		//if negated {
		//	p = "!" + p
		//}

		if negated {
			pos := len(out) - 1
			out[pos] = "!" + out[pos]

			if isBare {
				out[pos-1] = "!" + out[pos-1]
			}
		}
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

// DiscoverDocuments returns a list of all HTML and static files in the given root directory
//
// During discovery, excludes is a list of gitignore-style glob patterns
// which define what files must be entirely ignored during the discovery process.
// This means that any matching file is entirely excluded from discovery
// (and any subsequent processes which rely on it, such as document compilation).
// As a result, they will not be copied to the output directory at all.
//
// excludeCompile, on the other hand, is also a list of gitignore-style glob patterns.
// However, instead of entire exclusion, it simply reclassifies files as static
// so that future stages such as compilation will not attempt to compile them,
// but they still get copied to the output directory.
func DiscoverDocuments(root string, excludes []string, excludeCompile []string) (*DocumentLists, error) {
	lists := &DocumentLists{}

	excludes = append(defaultExcludes, excludes...)
	excludes = normaliseExcludes(excludes)
	//fmt.Println(excludes)
	excludePatterns, includePatterns := splitPatterns(excludes)

	excludeCompile = normaliseExcludes(excludeCompile)
	excludeCompilePatterns, includeCompilePatterns := splitPatterns(excludeCompile)

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
		if (ext == ".htm" || ext == ".html" || ext == ".shtml" || ext == ".xhtml") && !isExcluded(relPath, excludeCompilePatterns, includeCompilePatterns) {
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
