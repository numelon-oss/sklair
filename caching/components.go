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
	HeadNodes []*html.Node
	BodyNodes []*html.Node
	Dynamic   bool // whether the component (or any components contained within) contains any dynamic <lua> tags
}

type ComponentCache struct {
	Static  map[string]*Component
	Dynamic map[string]*Component
}

func MakeCache(source string, fileName string) (*Component, error) {
	path := filepath.Join(source, fileName)

	//if _, err := os.Stat(path); err != nil {
	//	return nil, err
	//}

	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// this is VERY naive but it actually works, we simply check for an opening lua tag
	// TODO: do the same check for html files
	hasLua := bytes.Contains(f, []byte("<lua"))
	component, err := html.Parse(bytes.NewReader(f))
	if err != nil {
		return nil, err
	}

	// even though components are usually bare (without doctype, head, body, etc), we still need to find the "body" (bc parsed)
	// because x/net/html automatically interprets the file as if its a full browser
	// ie it adds a doctype, head, body, etc tags automatically even if our input file doesnt have them

	//htmlNode := htmlUtilities.FindTag(component, "html")

	bodyNode := htmlUtilities.FindTag(component, "body")
	if bodyNode == nil {
		return nil, errors.New("no body tag found in component")
	}

	headNode := htmlUtilities.FindTag(component, "head")
	// we don't actually care about the head node, because it is not required,
	// however, x/net/html will automatically add a head anyway if not found in a component's source
	if headNode == nil {
		return nil, errors.New("no head tag found in component")
	}

	return &Component{
		HeadNodes: htmlUtilities.GetAllChildren(headNode), // TODO: in the final render of a source document, perform THOROUGH deduplication of head nodes
		BodyNodes: htmlUtilities.GetAllChildren(bodyNode),
		Dynamic:   hasLua, // TODO: when allowing circular components, Dynamic will be inherited based on whether it contains any components that are also dynamic
	}, nil
}
