{
  "description": "Parse Markdown once, render HTML, and inspect stable duplicate-safe headings through Sklair's Lua API."
}
---

# Markdown library

The `markdown` library parses source once and returns an immutable document object.

```lua
local document = markdown.parse(source)
local html = document:to_html()
local headings = document:headings()
local sections = document:sections()
```

`markdown.parse(source)` accepts exactly one string. The parser implements CommonMark with GitHub-flavoured extensions.

Inline code can opt into lexical highlighting by beginning with a hashtag and an exclamation mark, immediately followed by the language:

```markdown
`#!lua markdown.parse(source)`
```

This renders as `#!html <code class="language-lua">markdown.parse(source)</code>`; the marker is removed from rendered text, headings, generated identifiers, and extracted sections. Unmarked inline code retains ordinary CommonMark behaviour.

## `#!lua document:to_html(options?)`

Render the parsed document to a HTML fragment. By default, authored raw HTML is not passed through:

```lua
local safe = document:to_html()
local trusted = document:to_html({ rawHTML = true })
```

`rawHTML` is the only accepted option and must be boolean. Enable it only for content whose embedded HTML is trusted.

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
