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

var safeRenderer = goldmark.New(goldmark.WithExtensions(extension.GFM, inlineCodeExtension{}))
var rawRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM, inlineCodeExtension{}),
	goldmark.WithRendererOptions(goldmarkHTML.WithUnsafe()),
)

type Document struct {
	source   []byte
	root     ast.Node
	headings []Heading
	sections []Section
}

type Heading struct {
	Level int
	Title string
	ID    string
}

type Section struct {
	Level int
	Title string
	ID    string
	Text  string
}

func Parse(source string) *Document {
	content := []byte(source)
	root := safeRenderer.Parser().Parse(text.NewReader(content))
	document := &Document{source: content, root: root}
	document.headings = prepareHeadings(root, content)
	document.sections = prepareSections(root, content, document.headings)
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

func (d *Document) Sections() []Section {
	return append([]Section(nil), d.sections...)
}

func prepareSections(root ast.Node, source []byte, headings []Heading) []Section {
	sections := make([]Section, 0, len(headings))
	headingIndex := 0

	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		if _, ok := node.(*ast.Heading); ok {
			heading := headings[headingIndex]
			sections = append(sections, Section{
				Level: heading.Level,
				Title: heading.Title,
				ID:    heading.ID,
			})
			headingIndex++
			continue
		}
		if len(sections) == 0 {
			continue
		}

		text := plainText(node, source)
		if text == "" {
			continue
		}
		section := &sections[len(sections)-1]
		if section.Text != "" {
			section.Text += " "
		}
		section.Text += text
	}

	return sections
}

func plainText(root ast.Node, source []byte) string {
	var output strings.Builder
	appendText := func(value []byte) {
		if len(value) == 0 {
			return
		}
		if output.Len() > 0 {
			output.WriteByte(' ')
		}
		output.Write(value)
	}

	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch value := node.(type) {
		case *ast.CodeBlock:
			appendText(value.Text(source))
			return ast.WalkSkipChildren, nil
		case *ast.FencedCodeBlock:
			appendText(value.Text(source))
			return ast.WalkSkipChildren, nil
		case *ast.CodeSpan:
			text := codeSpanText(value, source)
			_, code, highlighted := highlightedCode(text)
			if highlighted {
				text = code
			}
			appendText([]byte(text))
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			appendText(value.Segment.Value(source))
		case *ast.String:
			appendText(value.Value)
		}
		return ast.WalkContinue, nil
	})

	return strings.Join(strings.Fields(output.String()), " ")
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
		title := headingText(heading, source)
		id := uniqueID(headingID(title), used)
		heading.SetAttribute([]byte("id"), []byte(id))
		headings = append(headings, Heading{Level: heading.Level, Title: title, ID: id})
		return ast.WalkContinue, nil
	})
	return headings
}

func headingText(root ast.Node, source []byte) string {
	var output strings.Builder
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch value := node.(type) {
		case *ast.CodeSpan:
			text := codeSpanText(value, source)
			_, code, highlighted := highlightedCode(text)
			if highlighted {
				text = code
			}
			output.WriteString(text)
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			output.Write(value.Segment.Value(source))
			if value.SoftLineBreak() || value.HardLineBreak() {
				output.WriteByte(' ')
			}
		case *ast.String:
			output.Write(value.Value)
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(output.String()), " ")
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
