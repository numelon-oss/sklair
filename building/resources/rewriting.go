package resources

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

func RewriteURLs(nodes []*html.Node, outputDir string) error {
	return Walk(nodes, func(reference Reference) error {
		var err error
		if reference.Attribute.Key == "srcset" {
			reference.Attribute.Val, err = rewriteSrcset(reference.Attribute.Val, outputDir)
		} else {
			reference.Attribute.Val, err = rewriteURL(reference.Attribute.Val, outputDir)
		}
		return err
	})
}

func rewriteSrcset(value string, outputDir string) (string, error) {
	candidates := strings.Split(value, ",")
	for i, candidate := range candidates {
		fields := strings.Fields(strings.TrimSpace(candidate))
		if len(fields) == 0 {
			continue
		}

		rewritten, err := rewriteURL(fields[0], outputDir)
		if err != nil {
			return "", err
		}
		fields[0] = rewritten
		candidates[i] = strings.Join(fields, " ")
	}

	return strings.Join(candidates, ", "), nil
}

func rewriteURL(value string, outputDir string) (string, error) {
	if value == "" || strings.HasPrefix(value, "/") || strings.HasPrefix(value, "#") || strings.HasPrefix(value, "//") {
		return value, nil
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return value, nil
	}

	if strings.HasPrefix(parsed.Path, "../") || parsed.Path == ".." {
		return "", fmt.Errorf("resource path %q escapes its resource directory", value)
	}

	cleaned := path.Clean(parsed.Path)
	if cleaned == "." {
		return value, nil
	}

	parsed.Path = path.Join(outputDir, cleaned)
	return parsed.String(), nil
}
