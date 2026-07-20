package building

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type renderRequest struct {
	layout string
	output string
	source string
	body   string
	props  map[string]sklairValue
	hook   string
	index  int
}

type generation struct {
	layout string
	body   string
	props  map[string]sklairValue
}

type documentQueue struct {
	layouts  map[string]string
	requests []renderRequest
	frozen   bool
}

func newDocumentQueue(layouts map[string]string) *documentQueue {
	return &documentQueue{layouts: layouts}
}

func (q *documentQueue) add(request renderRequest) error {
	if q.frozen {
		return fmt.Errorf("generated document queue is already frozen")
	}
	request.layout = strings.ToLower(strings.TrimSpace(request.layout))
	if request.layout == "" {
		return fmt.Errorf("layout cannot be empty")
	}
	if _, exists := q.layouts[request.layout]; !exists {
		return fmt.Errorf("layout %q does not exist", request.layout)
	}

	output, err := normaliseOutput(request.output)
	if err != nil {
		return err
	}
	request.output = output
	request.index = len(q.requests) + 1
	q.requests = append(q.requests, request)
	return nil
}

func normaliseOutput(output string) (string, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return "", fmt.Errorf("output cannot be empty")
	}
	if strings.Contains(output, "\\") {
		return "", fmt.Errorf("output %q must use / as its path separator", output)
	}
	if strings.ContainsRune(output, 0) {
		return "", fmt.Errorf("output contains a null byte")
	}
	if strings.Contains(output, ":") {
		return "", fmt.Errorf("output %q cannot contain a filesystem volume or URI scheme", output)
	}
	if path.IsAbs(output) {
		return "", fmt.Errorf("output %q must be relative to the build root", output)
	}
	if strings.HasSuffix(output, "/") {
		return "", fmt.Errorf("output %q must name a file", output)
	}

	normalised := path.Clean(output)
	if normalised == "." || normalised == ".." || strings.HasPrefix(normalised, "../") {
		return "", fmt.Errorf("output %q escapes the build root", output)
	}
	return normalised, nil
}

func (q *documentQueue) freeze(paths buildPaths, owners map[string]string) ([]plannedDocument, error) {
	if q.frozen {
		return nil, fmt.Errorf("generated document queue is already frozen")
	}
	q.frozen = true

	sort.SliceStable(q.requests, func(left int, right int) bool {
		return q.requests[left].output < q.requests[right].output
	})
	generated := make([]plannedDocument, 0, len(q.requests))
	for _, request := range q.requests {
		output := filepath.Join(paths.output, filepath.FromSlash(request.output))
		key := filepath.Clean(output)
		provenance := request.source
		description := fmt.Sprintf("generated request %d from hook %s", request.index, request.hook)
		if provenance != "" {
			description += " with source " + provenance
		} else {
			provenance = description
		}
		if previous, exists := owners[key]; exists {
			return nil, fmt.Errorf("output path %s is claimed by both %s and %s", output, previous, description)
		}
		owners[key] = description

		generated = append(generated, plannedDocument{
			plannedFile: plannedFile{source: provenance, output: output},
			generation: &generation{
				layout: request.layout,
				body:   request.body,
				props:  request.props,
			},
		})
	}
	return generated, nil
}
