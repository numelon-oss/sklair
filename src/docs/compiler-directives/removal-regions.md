{
  "description": "Exclude a region of the document head from compiled output with paired removal comments."
}
---

# Removal regions

Removal regions omit their contents and marker comments from the document head:

```html
<!-- sklair:remove -->
<meta name="development-note" content="never emitted">
<!-- sklair:remove-end -->
```

They are processed during [head segmentation](../sklair/head-processing.md), before deduplication and sorting. They cannot nest, and unmatched or unclosed markers fail compilation.

Place these directives in `head`. For conditional body content, use [props and `sklair-if`](../components/props-and-bindings.md) or [inline Lua](../lua/inline-syntax.md).
