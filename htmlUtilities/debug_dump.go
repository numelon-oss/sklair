package htmlUtilities

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func DumpNode(n *html.Node) {
	dumpNode(n, 0)
}

func DumpFromPoint(n *html.Node) {
	for c := n; c != nil; c = c.NextSibling {
		dumpNode(c, 0)
	}
}

func dumpNode(n *html.Node, depth int) {
	if n == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	switch n.Type {
	case html.ElementNode:
		fmt.Printf("%s<%s", indent, n.Data)
		for _, a := range n.Attr {
			fmt.Printf(` %s="%s"`, a.Key, a.Val)
		}
		fmt.Println(">")

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			fmt.Printf("%s  \"%s\"\n", indent, text)
		}

	case html.CommentNode:
		fmt.Printf("%s<!-- %s -->\n", indent, n.Data)

	default:
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		dumpNode(c, depth+1)
	}

	if n.Type == html.ElementNode {
		fmt.Printf("%s</%s>\n", indent, n.Data)
	}
}
