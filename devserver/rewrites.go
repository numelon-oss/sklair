package devserver

import (
	"fmt"
	"path"
	"strings"

	"sklair/sklairConfig"
)

type Rewrite struct {
	segments []string
	target   string
}

func CompileRewrites(values []sklairConfig.ServeRewrite) ([]Rewrite, error) {
	rewrites := make([]Rewrite, 0, len(values))
	patterns := make(map[string]struct{}, len(values))

	for index, value := range values {
		segments, err := rewritePattern(value.From)
		if err != nil {
			return nil, fmt.Errorf("serve rewrite %d: %w", index+1, err)
		}
		if _, exists := patterns[value.From]; exists {
			return nil, fmt.Errorf("serve rewrite %d: duplicate pattern %q", index+1, value.From)
		}
		patterns[value.From] = struct{}{}

		target, err := rewriteTarget(value.To)
		if err != nil {
			return nil, fmt.Errorf("serve rewrite %d: %w", index+1, err)
		}

		rewrites = append(rewrites, Rewrite{
			segments: segments,
			target:   target,
		})
	}

	return rewrites, nil
}

func rewritePattern(value string) ([]string, error) {
	if value == "" || !strings.HasPrefix(value, "/") {
		return nil, fmt.Errorf("pattern %q must begin with '/'", value)
	}
	if strings.ContainsAny(value, "?#") {
		return nil, fmt.Errorf("pattern %q cannot contain a query or fragment", value)
	}
	if path.Clean(value) != value {
		return nil, fmt.Errorf("pattern %q is not a clean path", value)
	}
	if value == "/" {
		return nil, nil
	}

	segments := strings.Split(strings.TrimPrefix(value, "/"), "/")
	for index, segment := range segments {
		switch {
		case segment == "*":
			if index != len(segments)-1 {
				return nil, fmt.Errorf("pattern %q can only use '*' as its final segment", value)
			}
		case strings.HasPrefix(segment, ":"):
			if !validParameterName(strings.TrimPrefix(segment, ":")) {
				return nil, fmt.Errorf("pattern %q contains an invalid named segment", value)
			}
		case strings.ContainsAny(segment, ":*"):
			return nil, fmt.Errorf("pattern %q contains an invalid segment", value)
		}
	}

	return segments, nil
}

func rewriteTarget(value string) (string, error) {
	if value == "" || !strings.HasPrefix(value, "/") {
		return "", fmt.Errorf("target %q must begin with '/'", value)
	}
	if strings.ContainsAny(value, "?#:*") {
		return "", fmt.Errorf("target %q must name one generated file", value)
	}
	if path.Clean(value) != value || value == "/" {
		return "", fmt.Errorf("target %q must be a clean file path", value)
	}

	return strings.TrimPrefix(value, "/"), nil
}

func validParameterName(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') &&
			character != '_' {
			return false
		}
	}
	return true
}

func (rewrite Rewrite) matches(requestPath string) bool {
	var segments []string
	if requestPath != "/" {
		if !strings.HasPrefix(requestPath, "/") {
			return false
		}
		segments = strings.Split(strings.TrimPrefix(requestPath, "/"), "/")
	}

	for index, pattern := range rewrite.segments {
		if pattern == "*" {
			return true
		}
		if index >= len(segments) || segments[index] == "" {
			return false
		}
		if strings.HasPrefix(pattern, ":") {
			continue
		}
		if pattern != segments[index] {
			return false
		}
	}

	return len(segments) == len(rewrite.segments)
}
