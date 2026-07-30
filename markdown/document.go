package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

var safeRenderer = goldmark.New(goldmark.WithExtensions(extension.GFM))
var rawRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(goldmarkHTML.WithUnsafe()),
)

type Document struct {
	source   []byte
	root     ast.Node
	headings []Heading
}

type Heading struct {
	Level int
	Title string
	ID    string
}

func Parse(source string) *Document {
	content := []byte(source)
	root := safeRenderer.Parser().Parse(text.NewReader(content))
	document := &Document{source: content, root: root}
	document.headings = prepareHeadings(root, content)
	return document
}

func (d *Document) HTML(rawHTML bool) (string, error) {
	var output bytes.Buffer
	renderer := safeRenderer.Renderer()
	if rawHTML {
		renderer = rawRenderer.Renderer()
	}
	if err := renderer.Render(&output, d.source, d.root); err != nil {
		return "", err
	}
	return output.String(), nil
}

func (d *Document) Headings() []Heading {
	return append([]Heading(nil), d.headings...)
}

func prepareHeadings(root ast.Node, source []byte) []Heading {
	headings := make([]Heading, 0)
	used := make(map[string]struct{})
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		title := strings.Join(strings.Fields(string(heading.Text(source))), " ")
		id := uniqueID(headingID(title), used)
		heading.SetAttribute([]byte("id"), []byte(id))
		headings = append(headings, Heading{Level: heading.Level, Title: title, ID: id})
		return ast.WalkContinue, nil
	})
	return headings
}

func headingID(title string) string {
	var id strings.Builder
	separator := false
	for _, character := range strings.ToLower(title) {
		switch {
		case unicode.IsLetter(character), unicode.IsNumber(character), unicode.IsMark(character):
			if separator && id.Len() > 0 {
				id.WriteByte('-')
			}
			separator = false
			id.WriteRune(character)
		case unicode.IsSpace(character), character == '-', character == '_':
			separator = true
		}
	}
	if id.Len() == 0 {
		return "heading"
	}
	return id.String()
}

func uniqueID(base string, used map[string]struct{}) string {
	if _, exists := used[base]; !exists {
		used[base] = struct{}{}
		return base
	}
	for suffix := 2; ; suffix++ {
		candidate := fmt.Sprintf("%s-%d", base, suffix)
		if _, exists := used[candidate]; !exists {
			used[candidate] = struct{}{}
			return candidate
		}
	}
}
