package building

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"sklair/htmlUtilities"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func prepareGenDocument(planned plannedDocument, static *staticLuaCompiler) (documentDefinition, error) {
	bodySource := []byte(planned.generation.body)
	if err := validateGenBodySource(bodySource); err != nil {
		return documentDefinition{}, fmt.Errorf("could not prepare generated document %s : %s", planned.source, err.Error())
	}

	context := &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"}
	nodes, err := html.ParseFragment(bytes.NewReader(bodySource), context)
	if err != nil {
		return documentDefinition{}, fmt.Errorf("could not parse generated document %s body : %s", planned.source, err.Error())
	}
	root := &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"}
	for _, node := range nodes {
		root.AppendChild(node)
	}

	dynamic, err := static.prepare(root, planned.source, false)
	if err != nil {
		return documentDefinition{}, fmt.Errorf("could not prepare generated document %s : %s", planned.source, err.Error())
	}
	if err := validateGenBodyNodes(htmlUtilities.GetAllChildren(root)); err != nil {
		return documentDefinition{}, fmt.Errorf("could not prepare generated document %s : %s", planned.source, err.Error())
	}

	return documentDefinition{
		plannedFile: planned.plannedFile,
		root:        root,
		lua:         dynamic,
		generation:  planned.generation,
	}, nil
}

func materialiseGenDocument(definition documentDefinition, layouts *layoutDefinitions) (documentDefinition, error) {
	body := htmlUtilities.Clone(definition.root)
	if err := bindDocument(body, definition.lua, definition.source, layouts.static); err != nil {
		return documentDefinition{}, err
	}
	if err := validateGenBodyNodes(htmlUtilities.GetAllChildren(body)); err != nil {
		return documentDefinition{}, fmt.Errorf("could not compile generated document %s : %s", definition.source, err.Error())
	}

	materialised, err := layouts.instantiate(
		definition.generation.layout,
		definition.generation.props,
		htmlUtilities.GetAllChildren(body),
		definition.plannedFile,
	)
	if err != nil {
		return documentDefinition{}, fmt.Errorf(
			"could not apply layout %s to generated document %s : %w",
			definition.generation.layout,
			definition.source,
			err,
		)
	}
	return materialised, nil
}

func validateGenBodySource(source []byte) error {
	tokeniser := html.NewTokenizer(bytes.NewReader(source))
	for {
		switch tokeniser.Next() {
		case html.ErrorToken:
			if err := tokeniser.Err(); err != nil && err != io.EOF {
				return err
			}
			return nil
		case html.DoctypeToken:
			return fmt.Errorf("generated document body cannot contain a doctype")
		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := tokeniser.TagName()
			if isDocShell(string(name)) {
				return fmt.Errorf("generated document body cannot contain <%s>", name)
			}
		}
	}
}

func validateGenBodyNodes(nodes []*html.Node) error {
	for _, node := range nodes {
		if node.Type == html.DoctypeNode {
			return fmt.Errorf("generated document body cannot contain a doctype")
		}
		if node.Type == html.ElementNode && isDocShell(node.Data) {
			return fmt.Errorf("generated document body cannot contain <%s>", node.Data)
		}
		if err := validateGenBodyNodes(htmlUtilities.GetAllChildren(node)); err != nil {
			return err
		}
	}
	return nil
}

func isDocShell(tag string) bool {
	switch strings.ToLower(tag) {
	case "html", "head", "body":
		return true
	default:
		return false
	}
}
