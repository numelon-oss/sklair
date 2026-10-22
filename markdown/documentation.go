package markdown

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type documentationExtension struct{}

func (documentationExtension) Extend(markdown goldmark.Markdown) {
	markdown.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(documentationRenderer{}, 500),
	))
}

type documentationRenderer struct{}

func (documentationRenderer) RegisterFuncs(registerer renderer.NodeRendererFuncRegisterer) {
	registerer.Register(ast.KindFencedCodeBlock, renderDocumentationCodeBlock)
	registerer.Register(ast.KindBlockquote, renderDocumentationBlockquote)
}

type codeFence struct {
	language  string
	title     string
	lineStart int
	highlight string
}

func renderDocumentationCodeBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	block := node.(*ast.FencedCodeBlock)
	if !entering {
		_, _ = writer.WriteString("</code></pre>\n")
		return ast.WalkContinue, nil
	}

	options, err := parseCodeFence(block, source)
	if err != nil {
		return ast.WalkStop, err
	}

	_, _ = writer.WriteString("<pre")
	if options.lineStart > 0 {
		_, _ = writer.WriteString(` class="line-numbers" data-start="`)
		_, _ = writer.WriteString(strconv.Itoa(options.lineStart))
		_ = writer.WriteByte('"')
	}
	if options.title != "" {
		_, _ = writer.WriteString(` data-title="`)
		_, _ = writer.WriteString(html.EscapeString(options.title))
		_ = writer.WriteByte('"')
	}
	if options.highlight != "" {
		_, _ = writer.WriteString(` data-line="`)
		_, _ = writer.WriteString(options.highlight)
		_ = writer.WriteByte('"')
	}
	_, _ = writer.WriteString("><code")
	if options.language != "" {
		_, _ = writer.WriteString(` class="language-`)
		_, _ = writer.WriteString(options.language)
		_ = writer.WriteByte('"')
	}
	_ = writer.WriteByte('>')
	for index := range block.Lines().Len() {
		line := block.Lines().At(index)
		_, _ = writer.Write(util.EscapeHTML(line.Value(source)))
	}
	return ast.WalkContinue, nil
}

func parseCodeFence(block *ast.FencedCodeBlock, source []byte) (codeFence, error) {
	if block.Info == nil {
		return codeFence{}, nil
	}
	fields, err := splitCodeFenceInfo(string(block.Info.Text(source)))
	if err != nil {
		return codeFence{}, err
	}
	if len(fields) == 0 {
		return codeFence{}, nil
	}

	result := codeFence{language: fields[0]}
	for _, character := range result.language {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '_' && character != '+' && character != '-' {
			return codeFence{}, fmt.Errorf("invalid fenced code language %q", result.language)
		}
	}

	seen := make(map[string]struct{})
	for _, field := range fields[1:] {
		name, value, present := strings.Cut(field, "=")
		if !present {
			if field != "linenums" {
				return codeFence{}, fmt.Errorf("invalid fenced code option %q", field)
			}
			name, value = field, "1"
		}
		if _, duplicate := seen[name]; duplicate {
			return codeFence{}, fmt.Errorf("duplicate fenced code option %q", name)
		}
		seen[name] = struct{}{}

		switch name {
		case "title":
			if value == "" {
				return codeFence{}, fmt.Errorf("fenced code title cannot be empty")
			}
			result.title = value
		case "linenums":
			if value == "true" {
				value = "1"
			}
			start, conversionError := strconv.Atoi(value)
			if conversionError != nil || start < 1 {
				return codeFence{}, fmt.Errorf("fenced code linenums must be a positive line number")
			}
			result.lineStart = start
		case "hl_lines":
			highlight, validationError := normaliseHighlightedLines(value)
			if validationError != nil {
				return codeFence{}, validationError
			}
			result.highlight = highlight
		default:
			return codeFence{}, fmt.Errorf("unknown fenced code option %q", name)
		}
	}
	return result, nil
}

func splitCodeFenceInfo(value string) ([]string, error) {
	fields := make([]string, 0)
	for index := 0; index < len(value); {
		for index < len(value) && unicode.IsSpace(rune(value[index])) {
			index++
		}
		if index == len(value) {
			break
		}

		var field strings.Builder
		quote := byte(0)
		for index < len(value) {
			character := value[index]
			if quote == 0 && unicode.IsSpace(rune(character)) {
				break
			}
			if character == '\'' || character == '"' {
				if quote == 0 {
					quote = character
					index++
					continue
				}
				if quote == character {
					quote = 0
					index++
					continue
				}
			}
			field.WriteByte(character)
			index++
		}
		if quote != 0 {
			return nil, fmt.Errorf("unterminated quote in fenced code options")
		}
		fields = append(fields, field.String())
	}
	return fields, nil
}

var highlightedLine = regexp.MustCompile(`^[0-9]+(?:-[0-9]+)?$`)

func normaliseHighlightedLines(value string) (string, error) {
	fields := strings.Fields(strings.ReplaceAll(value, ",", " "))
	if len(fields) == 0 {
		return "", fmt.Errorf("fenced code hl_lines cannot be empty")
	}
	for _, field := range fields {
		if !highlightedLine.MatchString(field) {
			return "", fmt.Errorf("invalid highlighted line %q", field)
		}
	}
	return strings.Join(fields, ","), nil
}

var admonitionStart = regexp.MustCompile(`^!!![ \t]+([a-zA-Z][a-zA-Z0-9-]*)(?:[ \t]+"([^"]*)")?[ \t]*$`)
var markdownFence = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})(.*)$")

const admonitionEnd = "<!-- sklair-admonition-end -->"

func preprocessAdmonitions(source string) string {
	lines := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	output := make([]string, 0, len(lines))
	var fenceCharacter byte
	var fenceLength int
	for index := 0; index < len(lines); index++ {
		fence := markdownFence.FindStringSubmatch(lines[index])
		if fenceCharacter != 0 {
			output = append(output, lines[index])
			if fence != nil && fence[1][0] == fenceCharacter && len(fence[1]) >= fenceLength && strings.TrimSpace(fence[2]) == "" {
				fenceCharacter = 0
			}
			continue
		}
		if fence != nil && (fence[1][0] != '`' || !strings.Contains(fence[2], "`")) {
			fenceCharacter = fence[1][0]
			fenceLength = len(fence[1])
			output = append(output, lines[index])
			continue
		}
		match := admonitionStart.FindStringSubmatch(lines[index])
		if match == nil || index+1 >= len(lines) {
			output = append(output, lines[index])
			continue
		}

		bodyStart := index + 1
		if lines[bodyStart] == "" {
			bodyStart++
		}
		if bodyStart >= len(lines) || (!strings.HasPrefix(lines[bodyStart], "    ") && !strings.HasPrefix(lines[bodyStart], "\t")) {
			output = append(output, lines[index])
			continue
		}

		title := match[2]
		if title == "" {
			title = strings.ToUpper(match[1][:1]) + match[1][1:]
		}
		output = append(output, "> [!"+strings.ToLower(match[1])+"] "+title, ">")
		index = bodyStart
		for ; index < len(lines); index++ {
			line := lines[index]
			switch {
			case strings.HasPrefix(line, "    "):
				output = append(output, "> "+line[4:])
			case strings.HasPrefix(line, "\t"):
				output = append(output, "> "+line[1:])
			case line == "":
				output = append(output, ">")
			default:
				index--
				goto admonitionComplete
			}
		}
		index--
	admonitionComplete:
		output = append(output, "", admonitionEnd, "")
	}
	return strings.Join(output, "\n")
}

var admonitionMarker = regexp.MustCompile(`^\[!([a-zA-Z][a-zA-Z0-9-]*)\](?:[ \t]+(.*))?$`)

func prepareAdmonitions(root ast.Node, source []byte) {
	var separators []*ast.HTMLBlock
	_ = ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if block, ok := node.(*ast.HTMLBlock); ok && strings.TrimSpace(string(block.Lines().Value(source))) == admonitionEnd {
			separators = append(separators, block)
			return ast.WalkSkipChildren, nil
		}
		blockquote, ok := node.(*ast.Blockquote)
		if !ok {
			return ast.WalkContinue, nil
		}
		first := blockquote.FirstChild()
		if first == nil || first.Kind() != ast.KindParagraph {
			return ast.WalkContinue, nil
		}
		match := admonitionMarker.FindStringSubmatch(strings.TrimSpace(string(first.Text(source))))
		if match == nil {
			return ast.WalkContinue, nil
		}
		blockquote.SetAttributeString("sklair-admonition", strings.ToLower(match[1]))
		blockquote.SetAttributeString("sklair-admonition-title", match[2])
		blockquote.RemoveChild(blockquote, first)
		return ast.WalkContinue, nil
	})
	for _, separator := range separators {
		separator.Parent().RemoveChild(separator.Parent(), separator)
	}
}

func renderDocumentationBlockquote(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	kindValue, admonition := node.AttributeString("sklair-admonition")
	if !admonition {
		if entering {
			_, _ = writer.WriteString("<blockquote>\n")
		} else {
			_, _ = writer.WriteString("</blockquote>\n")
		}
		return ast.WalkContinue, nil
	}

	kind := kindValue.(string)
	if entering {
		titleValue, _ := node.AttributeString("sklair-admonition-title")
		_, _ = writer.WriteString(`<aside class="documentation-admonition documentation-admonition--`)
		_, _ = writer.WriteString(html.EscapeString(kind))
		_, _ = writer.WriteString(`"><strong class="documentation-admonition__title">`)
		_, _ = writer.WriteString(html.EscapeString(titleValue.(string)))
		_, _ = writer.WriteString(`</strong><div class="documentation-admonition__body">`)
	} else {
		_, _ = writer.WriteString("</div></aside>\n")
	}
	return ast.WalkContinue, nil
}
