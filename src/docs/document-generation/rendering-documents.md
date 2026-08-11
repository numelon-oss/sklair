{
  "description": "Register generated documents from pre-build Lua with a layout, output path, structured props, body, and provenance."
}
---

# Rendering generated documents

`documents.render(request)` is available only in pre-build hooks. It accepts exactly one table.

```lua
documents.render({
  layout = "article",
  output = "journal/first-entry/index.html",
  source = "project:content/first-entry.md",
  props = {
    title = "First entry",
    tags = json.array({ "compiler", "html" }),
  },
  body = document:to_html(),
})
```

| Field | Required | Meaning |
| --- | --- | --- |
| `layout` | yes | discovered layout name |
| `output` | yes | final path relative to the output root |
| `source` | no | non-empty provenance label used in diagnostics |
| `props` | no | structured layout props |
| `body` | no | HTML fragment projected into the layout slot |

Unknown fields and non-string keys are errors. `layout` and `output` must be non-empty strings; `source`, when present, must also be non-empty. The body may be an empty string.

Props accept strings, numbers, booleans, arrays, objects, and explicit `json.null`. The top level must be an object, and prop names are normalised to lowercase without whitespace. Sklair snapshots the entire request at call time, including nested props, so later Lua mutations cannot alter an already queued document.

`source` is provenance, not an instruction to read a file. It makes an error say which input produced the request. The hook remains responsible for reading that input and supplying the compiled body.
