{
  "description": "Create a complete HTML layout, pass structured props, and insert generated content into its slot."
}
---

# Full-document layouts

Layouts are reusable complete HTML documents. They own the shell, metadata, navigation, assets, and one place where a generated body is projected.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title><span sklair-text="title"></span>  -  Documentation</title>
  </head>
  <body>
    <DocsNavigation></DocsNavigation>
    <main><slot></slot></main>
  </body>
</html>
```

Sklair discovers top-level HTML files in the configured layouts directory. The filename stem is the case-insensitive layout name: `layouts/Docs.html` is requested as `docs`. Nested layout directories are not discovered.

A layout must explicitly declare exactly one `html`, `head`, and `body`, and contain exactly one empty, attribute-free default `<slot>` inside the body. A head slot is forbidden. The strict shell makes ownership unambiguous: the generated body supplies content, while the layout supplies the document.

Layout props support structured values and are available to bindings and dynamic Lua. The generated body is compiled first in its own document scope with empty props, then projected into the instantiated layout. This prevents layout configuration from accidentally changing the meaning of the content body.

Static layout Lua runs once when a used layout definition is prepared. Dynamic layout Lua runs for every request. Components, head contributions, runtime-template requirements, and recursive resolution behave exactly as they do in source documents.
