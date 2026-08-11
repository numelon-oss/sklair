{
  "description": "Turn data and content into any number of complete HTML documents through Sklair's ordinary compiler."
}
---

# Document generation

Document generation lets one source process create any number of complete pages through Sklair's ordinary compiler. A directory of Markdown articles, an API description, or a product catalogue can become a routed site without a hook manually concatenating final HTML.

The model has three parts:

1. a pre-build hook reads and validates source material;
2. it registers document requests with `documents.render`;
3. Sklair compiles each body through a reusable full-document layout.

```lua
documents.render({
  layout = "docs",
  output = "docs/getting-started/index.html",
  source = "project:docs/getting-started.md",
  props = { title = "Getting started" },
  body = markdown.parse(source):to_html(),
})
```

Registration describes work; it does not write output immediately. Sklair validates and freezes the complete plan, detects collisions, then compiles generated documents with the same bindings, components, Lua, head finalisation, and runtime-template propagation as authored HTML documents.

Documentation sites are an excellent use of this feature, not a privileged one. The generator knows about layouts and documents. Your project decides whether those documents came from Markdown, JSON, schemas, databases exported in a pre-hook (somehow?), or something much more strange.
