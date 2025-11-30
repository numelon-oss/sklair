package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sklair/logger"
	"strings"

	"golang.org/x/net/html"
)

func clone(n *html.Node) *html.Node {
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
		clown.AppendChild(clone(child))
	}

	return clown
}

func main() {
	logger.InitShared(logger.LevelDebug)

	content, err := os.ReadFile("test.html")
	if err != nil {
		panic(err)
	}

	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		panic(err)
	}

	var toReplace []*html.Node

	for node := range doc.Descendants() {
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)

			if !htmlTags[tag] {
				toReplace = append(toReplace, node)
			}
		}
	}

	logger.Info("Found %d tags to replace", len(toReplace))

	for _, node := range toReplace {
		componentPath := filepath.Join("components", node.Data+".html")

		if _, err := os.Stat(componentPath); err != nil {
			logger.Error("Could not find component %s : %s", componentPath, err.Error())
			return
		} else {
			f, err := os.ReadFile(componentPath)
			if err != nil {
				logger.Error("Could not read component %s : %s", componentPath, err.Error())
				return
			}

			component, err := html.Parse(bytes.NewReader(f))
			if err != nil {
				logger.Error("Could not parse component %s : %s", componentPath, err.Error())
				return
			}

			// even though components are usually bare (without doctype, head, body, etc), we still need to find the "body" (bc parsed)
			body := component.FirstChild
			for body != nil && body.Data != "html" {
				body = body.NextSibling
			}
			if body != nil {
				body = body.FirstChild
				for body != nil && body.Data != "body" {
					body = body.NextSibling
				}
			}

			if body != nil {
				parent := node.Parent
				if parent != nil {
					for child := body.FirstChild; child != nil; child = child.NextSibling {
						parent.InsertBefore(clone(child), node)
					}
					parent.RemoveChild(node)
				}
			}
		}
	}

	newWriter := bytes.NewBuffer(nil)
	html.Render(newWriter, doc)

	os.WriteFile("output.html", newWriter.Bytes(), 0644)
	logger.Info("Saved to output.html")
}
