{
  "description": "Generate structured sites from Markdown and validated data without writing final HTML from hooks."
}
---

# Markdown-driven sites

Markdown-driven sites are where document generation becomes immediately obvious. The filesystem can provide information architecture, a small metadata header can provide page-specific details, and the first level-one heading can remain the canonical visible title.

A pre-build hook can scan content deterministically, parse each file, validate its optional metadata, and produce a document object:

```lua
local document = markdown.parse(markdown_source)
local headings = document:headings()

documents.render({
  layout = "documentation",
  output = route .. "/index.html",
  source = source_path,
  props = {
    description = metadata.description,
    navigation = navigation,
    headings = headings,
  },
  body = document:to_html(),
})
```

`markdown.parse(source)` returns an immutable document object. `document:headings()` exposes each heading's level, text, and generated identifier. `document:to_html()` uses safe HTML handling by default; pass `{ rawHTML = true }` only when authored raw HTML should be trusted.

Sklair's Markdown engine follows CommonMark with GitHub-flavoured extensions. Heading identifiers are Unicode-aware and duplicate-safe. The heading data is useful for local navigation, search indexes, or feeds, but none of those are compiler-mandated concepts.
