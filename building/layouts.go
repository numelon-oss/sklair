package building

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sklair/htmlUtilities"
	"strings"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
)

type layoutDefinition struct {
	source string
	root   *html.Node
	lua    *dynamicLuaDefinition
}

type layoutDefinitions struct {
	root     string
	sources  map[string]string
	prepared map[string]*layoutDefinition
	static   *staticLuaCompiler
}

func (d *layoutDefinitions) prepare(name string) (*layoutDefinition, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	source, exists := d.sources[name]
	if !exists {
		return nil, fmt.Errorf("layout %q does not exist", name)
	}
	if definition, exists := d.prepared[name]; exists {
		return definition, nil
	}

	logger.Info("Preparing layout %s...", name)
	path := filepath.Join(d.root, source)
	definition, err := prepareLayout(path, d.static)
	if err != nil {
		return nil, fmt.Errorf("could not prepare layout %s : %s", source, err.Error())
	}
	d.prepared[name] = definition
	return definition, nil
}

func prepareLayout(path string, static *staticLuaCompiler) (*layoutDefinition, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := naiveValidation(source, "layout"); err != nil {
		return nil, err
	}
	if err := validateLayoutDocument(source); err != nil {
		return nil, err
	}

	root, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}
	dynamic, err := static.prepare(root, path, false)
	if err != nil {
		return nil, err
	}

	head := htmlUtilities.FindTag(root, "head")
	body := htmlUtilities.FindTag(root, "body")
	if head == nil || body == nil {
		return nil, errors.New("layout must be a complete document with head and body tags")
	}
	if err := validateLayoutSlot(head, body); err != nil {
		return nil, err
	}
	if err := rejectDataSklairRuntime(htmlUtilities.GetAllChildren(root)); err != nil {
		return nil, err
	}

	return &layoutDefinition{source: path, root: root, lua: dynamic}, nil
}

func validateLayoutDocument(source []byte) error {
	found := make(map[string]bool)
	tokeniser := html.NewTokenizer(bytes.NewReader(source))
	for {
		switch tokeniser.Next() {
		case html.ErrorToken:
			if err := tokeniser.Err(); err != nil && err != io.EOF {
				return err
			}
			for _, tag := range []string{"html", "head", "body"} {
				if !found[tag] {
					return fmt.Errorf("layout does not explicitly declare <%s>", tag)
				}
			}
			return nil
		case html.StartTagToken:
			name, _ := tokeniser.TagName()
			found[strings.ToLower(string(name))] = true
		}
	}
}

func validateLayoutSlot(head *html.Node, body *html.Node) error {
	if len(findSlots(htmlUtilities.GetAllChildren(head))) > 0 {
		return errors.New("layout slots are not allowed in <head>")
	}

	slots := findSlots(htmlUtilities.GetAllChildren(body))
	if len(slots) == 0 {
		return errors.New("layout does not declare a default slot")
	}
	if len(slots) > 1 {
		return errors.New("layout body contains more than one slot")
	}
	if len(slots[0].Attr) > 0 {
		return errors.New("default slot attributes are not supported")
	}
	if htmlUtilities.HasChildren(slots[0]) {
		return errors.New("default slot fallback content is not supported")
	}
	return nil
}

func (d *layoutDefinitions) instantiate(name string, props map[string]sklairValue, projectedBody []*html.Node, planned plannedFile) (documentDefinition, error) {
	definition, err := d.prepare(name)
	if err != nil {
		return documentDefinition{}, err
	}

	root := htmlUtilities.Clone(definition.root)
	head := htmlUtilities.FindTag(root, "head")
	body := htmlUtilities.FindTag(root, "body")
	if head == nil || body == nil {
		return documentDefinition{}, fmt.Errorf("could not find head or body tags in layout %s after cloning", name)
	}

	bound, err := bindValues(htmlUtilities.GetAllChildren(root), props, definition.lua, "layout")
	if err != nil {
		return documentDefinition{}, fmt.Errorf("could not bind layout %s : %s", name, err.Error())
	}
	if err := d.static.runDynamic([]*html.Node{root}, definition.lua, bound, definition.source); err != nil {
		return documentDefinition{}, err
	}
	if err := bound.normalise(); err != nil {
		return documentDefinition{}, fmt.Errorf("could not normalise layout %s props : %s", name, err.Error())
	}

	if err := validateLayoutSlot(head, body); err != nil {
		return documentDefinition{}, fmt.Errorf("layout %s default slot is not available for these props : %s", name, err.Error())
	}
	slot := findSlot(htmlUtilities.GetAllChildren(body))
	htmlUtilities.InsertNodesBefore(slot, projectedBody)
	slot.Parent.RemoveChild(slot)

	return documentDefinition{plannedFile: planned, root: root}, nil
}
