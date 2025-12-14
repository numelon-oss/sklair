package caching

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sklair/htmlUtilities"

	"golang.org/x/net/html"
)

type Component struct {
	Nodes   []*html.Node
	Dynamic bool // whether the component contains any dynamic <lua> tags
}

type ComponentCache struct {
	Static  map[string]*Component
	Dynamic map[string]*Component
}

func Cache(source string, fileName string) (*Component, bool, error) {
	path := filepath.Join(source, fileName)

	//if _, err := os.Stat(path); err != nil {
	//	return nil, err
	//}

	f, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	// this is VERY naive but it actually works, we simply check for an opening lua tag
	// TODO: do the same check for html files
	hasLua := bytes.Contains(f, []byte("<lua"))
	component, err := html.Parse(bytes.NewReader(f))
	if err != nil {
		return nil, false, err
	}

	// even though components are usually bare (without doctype, head, body, etc), we still need to find the "body" (bc parsed)
	// because x/net/html automatically interprets the file as if its a full browser
	// ie it adds a doctype, head, body, etc tags automatically even if our input file doesnt have them

	htmlNode := htmlUtilities.FindTag(component, "html")
	bodyNode := htmlUtilities.FindTag(htmlNode, "body")

	if bodyNode == nil {
		return nil, false, errors.New("no body tag found in component")
	}

	var children []*html.Node
	for child := bodyNode.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}

	// TODO: make a new struct for components which includes a head section and a body section
	// for head, perform deduplication when multiple components in same document share head stuff
	// for body, just insert as usual

	//fmt.Println(children)
	return &Component{children, hasLua}, hasLua, nil
}
