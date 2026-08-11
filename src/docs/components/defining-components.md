{
  "description": "Define a Sklair component and invoke it from a HTML source document or another component."
}
---

# Defining components

The configured components directory is a flat namespace. Each HTML file defines one component whose name is the filename without its extension:

```text
components/
├── SiteHeader.html
└── ReleaseCard.html
```

`<SiteHeader></SiteHeader>` resolves to `SiteHeader.html`. Names are case-insensitive, so `<siteheader>` resolves to the same definition. Two entries which collapse to the same lowercase name are an error rather than a coin toss.

## Definition shape

A complete definition makes its two responsibilities explicit:

```html
<!doctype html>
<html>
<head>
  <link rel="stylesheet" href="/styles/release-card.css">
</head>
<body>
  <article class="release-card">
    <h2 sklair-text="title"></h2>
  </article>
</body>
</html>
```

Nodes from the body replace the component invocation. Nodes from the head are contributed to the containing document and enter normal final deduplication and ordering.

HTML fragments are accepted by the parser and receive an implicit document shell, but explicit `<head>` and `<body>` tags make ownership considerably easier to read - especially once a component gains dependencies.

## Invocation

```html
<ReleaseCard title="Version 2.4"></ReleaseCard>
```

Component attributes become compile-time props. Child nodes are rejected unless the definition declares a default slot.

Native HTML elements always remain native. An unknown non-standard element which has no matching component is preserved as an [autonomous custom element](https://developer.mozilla.org/docs/Web/API/Web_components/Using_custom_elements) and produces a warning.

## Lazy preparation

Discovery records component names and entry paths. Sklair parses a definition only when a source document, generated document, layout, or another component reaches it. The prepared definition is cached for the build.

This matters for static Lua and large component libraries: used definitions perform their work once, while unused definitions do no work at all.
