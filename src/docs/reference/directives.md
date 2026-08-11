{
  "description": "Reference Sklair's source directives for head-ordering barriers and compile-time removal regions."
}
---

# Compiler directives

Sklair recognises two paired HTML-comment directives. They are compiler syntax and disappear with the source structure they control.

## Ordering barriers

An ordering barrier keeps several head nodes together while telling the head optimiser where the block belongs:

```html
<!-- sklair:ordering-barrier treat-as=stylesheet -->
<link rel="stylesheet" href="/fonts.css">
<style>.icon { font-family: Sklair Icons; }</style>
<!-- sklair:ordering-barrier-end -->
```

Allowed `treat-as` values are:

- `charset`
- `http-equiv`
- `viewport`
- `resource-hint`
- `title`
- `stylesheet`
- `script`
- `analytics`
- `dont-care`

The complete block is ordered at that priority. Its internal node order remains intact, and nodes inside it are not individually deduplicated. Barriers cannot nest. A missing, unknown, unmatched, or unclosed directive fails compilation.

Use barriers when two head nodes have a real ordering relationship which ordinary per-element classification cannot express. Sklair orders explicit resource hints; it does not invent preload or preconnect tags automatically (big **yet**).

## Removal regions

```html
<!-- sklair:remove -->
<meta name="development-note" content="never emitted">
<!-- sklair:remove-end -->
```

Removal regions are processed while segmenting the document head. Their contents and marker comments are omitted from output. They cannot nest, and unmatched or unclosed markers fail compilation.

Because removal is a head-segmentation feature, place these directives in `head`. They are not a general conditional system for body content; use props, `sklair-if`, or compile-time Lua there.
