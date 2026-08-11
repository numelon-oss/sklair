{
  "description": "Understand generated output ownership, normalisation, deterministic ordering, collisions, and source provenance."
}
---

# Generated output planning

Generated requests join authored documents and static files in one output plan. Sklair freezes that plan after pre-build hooks finish, before compilation changes the output directory.

Output paths must be clean relative paths. Absolute paths, drive or namespace colons, backslashes, traversal, null bytes, and trailing directory separators are rejected. Use **forward slashes**, *regardless of the host operating system*.

Sklair normalises and sorts generated requests before compilation. Every final path has one owner: two generated documents cannot claim the same path, and a generated document cannot overwrite an authored document or copied static file. The collision fails before replacement of the previous successful output.

```lua
documents.render({
  layout = "article",
  output = "docs/compiler/index.html",
  source = "project:articles/compiler.md",
  body = html,
})
```

The optional source value appears in diagnostics and collision reports. Use a namespaced value such as `project:articles/compiler.md` when that is the real provenance, but remember that Sklair never opens it on the request's behalf.
