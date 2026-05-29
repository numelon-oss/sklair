package htmlUtilities

import (
	"fmt"
	"hash/maphash"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func Clone(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}

	// hah
	clown := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     append([]html.Attribute{}, n.Attr...),
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		clown.AppendChild(Clone(child))
	}

	return clown
}

func FindTag(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := FindTag(c, tag); found != nil {
			return found
		}
	}

	return nil
}

func HasChildren(n *html.Node) bool {
	return n != nil && n.FirstChild != nil
}

func InsertNodesBefore(insertBefore *html.Node, tags []*html.Node) {
	for _, tag := range tags {
		insertBefore.Parent.InsertBefore(Clone(tag), insertBefore)
	}
}

func AppendNodes(parent *html.Node, tags []*html.Node) {
	for _, tag := range tags {
		parent.AppendChild(Clone(tag))
	}
}

func GetAllChildren(n *html.Node) []*html.Node {
	var children []*html.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}

	return children
}

var seedOfInnovation = maphash.MakeSeed()

func superweakHash(s string) uint64 {
	var h maphash.Hash
	h.SetSeed(seedOfInnovation)
	_, _ = h.WriteString(s)
	return h.Sum64()
}

func WeakHashNode(n *html.Node) uint64 {
	if n == nil {
		return 0
	}

	var lala strings.Builder

	lala.WriteString(strconv.Itoa(int(n.Type)))
	lala.WriteString("|")
	lala.WriteString(n.Data)
	lala.WriteString("|")

	for _, attr := range n.Attr {
		lala.WriteString(attr.Key + "=" + attr.Val)
		lala.WriteString(";")
	}

	// only consider input text for script and style tags
	if (n.Data == "script" || n.Data == "style") && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		lala.WriteString("|")
		lala.WriteString(n.FirstChild.Data)
	}

	return superweakHash(lala.String())
}

func CommentNode(comment string) *html.Node {
	return &html.Node{Type: html.CommentNode, Data: comment}
}

func RemoveAllChildren(n *html.Node) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		n.RemoveChild(c)
		c = next
	}
}

func ParseFile(filePath string) (*html.Node, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filePath, err)
	}

	return doc, nil
}
