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

// naiveValidation has one purpose:
// if you tried to use this feature, you must at least LOOK like you used it correctly,
// otherwise later stages will come back to bite you
func naiveValidation(b []byte) error {
	// TODO: ensure these are only naively detected within comments
	if bytes.Contains(b, []byte("sklair:ordering-barrier")) {
		if !bytes.Contains(b, []byte("treat-as=")) {
			return errors.New("ordering barrier missing treat-as= in component")
		}
		if !bytes.Contains(b, []byte("sklair:ordering-barrier-end")) {
			return errors.New("unterminated ordering barrier in component")
		}
	}

	if bytes.Contains(b, []byte("sklair:remove")) {
		if !bytes.Contains(b, []byte("sklair:remove-end")) {
			return errors.New("unterminated remove directive in component")
		}
	}

	return nil
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

	// the idea is that we want MakeCache to CHEAPLY validate things (instead of implementing a mini parser here),
	// and any actual errors will be caught in later stages
	// (like parsing)
	veryVeryNaive := naiveValidation(f)
	if veryVeryNaive != nil {
		return nil, veryVeryNaive
	}

	// this is VERY naive, but it actually works; we simply check for an opening lua tag
	hasLua := bytes.Contains(f, []byte("<lua"))
	component, err := html.Parse(bytes.NewReader(f))
	if err != nil {
		return nil, err
	}

	// even though components are usually bare (without doctype, head, body, etc.),
	// we still need to find the "body" (bc parsed)
	// because x/net/html automatically interprets the file just like a full browser would.
	// i.e. it adds a doctype, head, body, etc. tags automatically even if our input file doesn't have them

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
		HeadNodes: htmlUtilities.GetAllChildren(headNode),
		BodyNodes: htmlUtilities.GetAllChildren(bodyNode),
		Dynamic:   hasLua, // TODO: when allowing circular components, Dynamic will be inherited based on whether it contains any components that are also dynamic
	}, nil
}
