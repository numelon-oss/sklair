package building

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sklair/discovery"
	"sklair/htmlUtilities"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
)

type documentDefinition struct {
	plannedFile
	root *html.Node
}

type componentDefinition struct {
	head    []*html.Node
	body    []*html.Node
	dynamic bool
}

type componentDefinitions struct {
	root     string
	sources  map[string]discovery.ComponentSource
	prepared map[string]*componentDefinition
}

type definitionSet struct {
	documents  []documentDefinition
	components *componentDefinitions
}

func prepareDefinitions(inputs *buildInputs, plan *buildPlan) (*definitionSet, error) {
	documents := make([]documentDefinition, 0, len(plan.documents))
	for _, planned := range plan.documents {
		root, err := htmlUtilities.ParseFile(planned.source)
		if err != nil {
			return nil, fmt.Errorf("could not parse file %s : %s", planned.source, err.Error())
		}

		documents = append(documents, documentDefinition{
			plannedFile: planned,
			root:        root,
		})
	}

	return &definitionSet{
		documents: documents,
		components: &componentDefinitions{
			root:     inputs.paths.components,
			sources:  inputs.components,
			prepared: make(map[string]*componentDefinition),
		},
	}, nil
}

func (d *componentDefinitions) prepare(name string) (*componentDefinition, discovery.ComponentSource, error) {
	source, exists := d.sources[name]
	if !exists {
		return nil, discovery.ComponentSource{}, fmt.Errorf("component %q does not exist", name)
	}

	if definition, exists := d.prepared[name]; exists {
		return definition, source, nil
	}

	logger.Info("Preparing component %s...", name)
	definition, err := prepareComponent(filepath.Join(d.root, source.Entry()))
	if err != nil {
		return nil, discovery.ComponentSource{}, fmt.Errorf("could not prepare component %s : %s", source.Entry(), err.Error())
	}

	d.prepared[name] = definition
	return definition, source, nil
}

func prepareComponent(path string) (*componentDefinition, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := naiveValidation(source); err != nil {
		return nil, err
	}

	root, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	head := htmlUtilities.FindTag(root, "head")
	if head == nil {
		return nil, errors.New("no head tag found in component")
	}

	body := htmlUtilities.FindTag(root, "body")
	if body == nil {
		return nil, errors.New("no body tag found in component")
	}

	return &componentDefinition{
		head:    htmlUtilities.GetAllChildren(head),
		body:    htmlUtilities.GetAllChildren(body),
		dynamic: bytes.Contains(source, []byte("<lua")),
	}, nil
}

// naiveValidation has one purpose:
// if you tried to use this feature, you must at least LOOK like you used it correctly,
// otherwise later stages will come back to bite you
func naiveValidation(source []byte) error {
	// TODO: ensure these are only naively detected within comments
	if bytes.Contains(source, []byte("sklair:ordering-barrier")) {
		if !bytes.Contains(source, []byte("treat-as=")) {
			return errors.New("ordering barrier missing treat-as= in component")
		}
		if !bytes.Contains(source, []byte("sklair:ordering-barrier-end")) {
			return errors.New("unterminated ordering barrier in component")
		}
	}

	if bytes.Contains(source, []byte("sklair:remove")) && !bytes.Contains(source, []byte("sklair:remove-end")) {
		return errors.New("unterminated remove directive in component")
	}

	return nil
}
