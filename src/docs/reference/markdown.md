{
  "description": "Parse Markdown, render HTML, inspect headings and links, and use code options and admonitions."
}
---

# Markdown library

The `markdown` library parses source once and returns an immutable document object.

```lua
local document = markdown.parse(source)
local html = document:to_html()
local headings = document:headings()
local links = document:links()
local sections = document:sections()
```

`markdown.parse(source)` accepts exactly one string. The parser implements CommonMark with GitHub-flavoured extensions.

Inline code can opt into lexical highlighting by beginning with a hashtag and an exclamation mark, immediately followed by the language:

```markdown
`#!lua markdown.parse(source)`
```

This renders as `#!html <code class="language-lua">markdown.parse(source)</code>`; the marker is removed from rendered text, headings, generated identifiers, and extracted sections. Unmarked inline code retains ordinary CommonMark behaviour.

## Fenced code options

Fenced code blocks accept an optional title, starting line number, and highlighted lines:

````markdown title="Markdown source"
```lua title="hooks/pre/articles.lua" linenums="12" hl_lines="14 17-19"
local content = fs.read("project:articles/introduction.md")
```
````

`linenums` may be written without a value to begin at line 1. `hl_lines` accepts individual lines and inclusive ranges separated by spaces or commas.

## Admonitions

Use an admonition when information should not disappear into the surrounding paragraphs:

```markdown title="Markdown source"
!!! warning "Trusted HTML"

    Only pass HTML which you control or have deliberately sanitised.
```

The type becomes a `documentation-admonition--<type>` class. This documentation uses:

| Type | Colour | Use |
| --- | --- | --- |
| `info` | Blue | Additional information or context |
| `tip` | Green | Advice and shortcuts |
| `warning` | Amber | Behaviour that needs attention |
| `critical` | Red | Unsafe input, possible data loss, or other serious consequences |

`note` uses the info styling and `danger` uses the critical styling. The whole box is tinted, with a stronger title background and matching border. Custom types are also accepted; generated sites can supply their own CSS.

## `#!lua document:to_html(options?)`

Render the parsed document to a HTML fragment. By default, authored raw HTML is not passed through:

```lua
local safe = document:to_html()
local trusted = document:to_html({ rawHTML = true })
```

Enable `rawHTML` only for content whose embedded HTML is trusted.

The optional `links` table replaces exact Markdown link destinations while rendering:

```lua
local html = document:to_html({
  links = {
    ["../guide.md"] = "/docs/guide/",
  },
})
```

This is useful when source-relative Markdown paths and generated output URLs follow different layouts. Images are not rewritten.

## `#!lua document:links()`

Return the document's unique Markdown link destinations as a sorted array. A generator can resolve them against its source map and pass the replacements to `to_html`.

## `#!lua document:headings()`

Return an ordered array of every heading:

```lua
{
  { level = 2, title = "Installation", id = "installation" },
  { level = 2, title = "Installation", id = "installation-1" },
}
```

Heading identifiers are stable, Unicode-aware, and made unique by suffixing duplicates. The same identifiers are used in rendered HTML, so a navigation tree built from `headings()` links to the correct elements.

## `#!lua document:sections()`

Return one non-overlapping search-oriented section for every heading. Each item contains `level`, `title`, `id`, and the plain textual content between that heading and the next heading:

```lua
{
  {
    level = 2,
    title = "Runtime bindings",
    id = "runtime-bindings",
    text = "Template source uses the same spelling as compile-time bindings...",
  },
}
```

Text is derived from the parsed Markdown tree rather than by stripping syntax from the original source. Code blocks remain searchable, while structural Markdown punctuation and raw HTML tags do not become search terms.
