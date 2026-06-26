package building

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"strings"

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
	if err := validateSlots(htmlUtilities.GetAllChildren(head), htmlUtilities.GetAllChildren(body)); err != nil {
		return nil, err
	}

	return &componentDefinition{
		head:    htmlUtilities.GetAllChildren(head),
		body:    htmlUtilities.GetAllChildren(body),
		dynamic: bytes.Contains(source, []byte("<lua")),
	}, nil
}

func validateSlots(head []*html.Node, body []*html.Node) error {
	if len(findSlots(head)) > 0 {
		return errors.New("component slots are not allowed in <head>")
	}

	slots := findSlots(body)
	if len(slots) > 1 {
		return errors.New("component body contains more than one slot")
	}
	if len(slots) == 0 {
		return nil
	}

	slot := slots[0]
	if len(slot.Attr) > 0 {
		return errors.New("default slot attributes are not supported")
	}
	if htmlUtilities.HasChildren(slot) {
		return errors.New("default slot fallback content is not supported")
	}

	return nil
}

func findSlots(nodes []*html.Node) []*html.Node {
	var slots []*html.Node
	for _, node := range nodes {
		collectSlots(node, &slots)
	}
	return slots
}

func collectSlots(node *html.Node, slots *[]*html.Node) {
	if node.Type == html.ElementNode && node.Data == "slot" {
		*slots = append(*slots, node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectSlots(child, slots)
	}
}

func findSlot(nodes []*html.Node) *html.Node {
	slots := findSlots(nodes)
	if len(slots) == 0 {
		return nil
	}
	return slots[0]
}

func validateComponentMode(definition *componentDefinition, runtimeTemplate bool) error {
	if err := rejectDataSklairRuntime(definition.head); err != nil {
		return err
	}
	if err := rejectDataSklairRuntime(definition.body); err != nil {
		return err
	}
	if runtimeTemplate {
		if binding := findSklairBinding(definition.head); binding != "" {
			return fmt.Errorf("runtime binding %q is not allowed in component <head>", binding)
		}
	}
	return nil
}

func rejectDataSklairRuntime(nodes []*html.Node) error {
	for _, node := range nodes {
		if node.Type == html.ElementNode {
			for _, attribute := range node.Attr {
				if strings.HasPrefix(attribute.Key, "data-sklair-") {
					return fmt.Errorf("%q is generated runtime syntax; write %q in component source", attribute.Key, strings.TrimPrefix(attribute.Key, "data-"))
				}
			}
		}
		if err := rejectDataSklairRuntime(htmlUtilities.GetAllChildren(node)); err != nil {
			return err
		}
	}
	return nil
}

func findSklairBinding(nodes []*html.Node) string {
	for _, node := range nodes {
		if node.Type == html.ElementNode {
			for _, attribute := range node.Attr {
				if strings.HasPrefix(attribute.Key, "sklair-") {
					return attribute.Key
				}
			}
		}
		if binding := findSklairBinding(htmlUtilities.GetAllChildren(node)); binding != "" {
			return binding
		}
	}
	return ""
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
