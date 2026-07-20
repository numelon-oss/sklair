package building

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type plannedFile struct {
	source string
	output string
}

type plannedDocument struct {
	plannedFile
	generation *generation
}

type buildPlan struct {
	documents   []plannedDocument
	staticFiles []plannedFile
}

func planBuild(inputs *buildInputs, queue *documentQueue) (*buildPlan, error) {
	owners := make(map[string]string, len(inputs.documents.HtmlFiles)+len(inputs.documents.StaticFiles))

	sourceDocuments, err := planFiles(inputs.documents.HtmlFiles, inputs.paths, owners)
	if err != nil {
		return nil, err
	}

	staticFiles, err := planFiles(inputs.documents.StaticFiles, inputs.paths, owners)
	if err != nil {
		return nil, err
	}
	generated, err := queue.freeze(inputs.paths, owners)
	if err != nil {
		return nil, err
	}

	documents := make([]plannedDocument, 0, len(sourceDocuments)+len(generated))
	for _, source := range sourceDocuments {
		documents = append(documents, plannedDocument{plannedFile: source})
	}
	documents = append(documents, generated...)
	sort.Slice(documents, func(left int, right int) bool {
		return documents[left].output < documents[right].output
	})

	return &buildPlan{
		documents:   documents,
		staticFiles: staticFiles,
	}, nil
}

func (p *buildPlan) generatedCount() int {
	count := 0
	for _, document := range p.documents {
		if document.generation != nil {
			count++
		}
	}
	return count
}

func planFiles(files []string, paths buildPaths, owners map[string]string) ([]plannedFile, error) {
	planned := make([]plannedFile, 0, len(files))

	for _, source := range files {
		rel, err := filepath.Rel(paths.input, source)
		if err != nil {
			return nil, fmt.Errorf("could not get relative path for %s : %s", source, err.Error())
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("source path %s is outside the input directory", source)
		}

		output := filepath.Join(paths.output, rel)
		key := filepath.Clean(output)
		if previous, exists := owners[key]; exists {
			return nil, fmt.Errorf("output path %s is claimed by both %s and %s", output, previous, source)
		}
		owners[key] = source

		planned = append(planned, plannedFile{
			source: source,
			output: output,
		})
	}

	return planned, nil
}
