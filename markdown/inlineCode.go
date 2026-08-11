package markdown

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type inlineCodeExtension struct{}

func (inlineCodeExtension) Extend(markdown goldmark.Markdown) {
	markdown.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(inlineCodeRenderer{}, 500),
	))
}

type inlineCodeRenderer struct{}

func (inlineCodeRenderer) RegisterFuncs(registerer renderer.NodeRendererFuncRegisterer) {
	registerer.Register(ast.KindCodeSpan, renderInlineCode)
}

func renderInlineCode(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	value := codeSpanText(node.(*ast.CodeSpan), source)
	language, code, highlighted := highlightedCode(value)
	if highlighted {
		_, _ = writer.WriteString(`<code class="language-`)
		_, _ = writer.WriteString(language)
		_, _ = writer.WriteString(`">`)
		_, _ = writer.Write(util.EscapeHTML([]byte(code)))
		_, _ = writer.WriteString(`</code>`)
		return ast.WalkSkipChildren, nil
	}

	_, _ = writer.WriteString(`<code>`)
	_, _ = writer.Write(util.EscapeHTML([]byte(value)))
	_, _ = writer.WriteString(`</code>`)
	return ast.WalkSkipChildren, nil
}

func codeSpanText(node *ast.CodeSpan, source []byte) string {
	var value bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		text := child.(*ast.Text).Segment.Value(source)
		if bytes.HasSuffix(text, []byte("\n")) {
			value.Write(text[:len(text)-1])
			value.WriteByte(' ')
			continue
		}
		value.Write(text)
	}
	return value.String()
}

func highlightedCode(value string) (string, string, bool) {
	if !strings.HasPrefix(value, "#!") {
		return "", value, false
	}

	separator := strings.IndexFunc(value[2:], unicode.IsSpace)
	if separator < 1 {
		return "", value, false
	}
	separator += 2
	language := value[2:separator]
	for _, character := range []byte(language) {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') &&
			character != '_' && character != '+' && character != '-' {
			return "", value, false
		}
	}

	code := strings.TrimLeftFunc(value[separator:], unicode.IsSpace)
	if code == "" {
		return "", value, false
	}

	return language, code, true
}
