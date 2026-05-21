package resources

import "golang.org/x/net/html"

type Reference struct {
	Node      *html.Node
	Attribute *html.Attribute
}

type Visitor func(Reference) error

func Walk(nodes []*html.Node, visitor Visitor) error {
	for _, node := range nodes {
		if err := walkNode(node, visitor); err != nil {
			return err
		}
	}

	return nil
}

func walkNode(node *html.Node, visitor Visitor) error {
	if node.Type == html.ElementNode {
		for i := range node.Attr {
			if !isReferenceAttribute(node.Data, node.Attr[i].Key) {
				continue
			}

			if err := visitor(Reference{
				Node:      node,
				Attribute: &node.Attr[i],
			}); err != nil {
				return err
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := walkNode(child, visitor); err != nil {
			return err
		}
	}

	return nil
}

func isReferenceAttribute(tag string, attribute string) bool {
	switch tag {
	case "link":
		return attribute == "href"
	case "script", "img", "source", "video", "audio", "track", "embed", "iframe":
		return attribute == "src" || attribute == "srcset"
	case "object":
		return attribute == "data"
	}

	return false
}
