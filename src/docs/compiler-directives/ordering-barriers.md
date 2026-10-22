{
  "description": "Keep related head nodes in their source order and assign the group a head-processing priority."
}
---

# Ordering barriers

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

The complete block is ordered at that priority. Its internal node order remains intact, and nodes inside it are not individually deduplicated. Barriers cannot nest. An unknown type, unmatched end, or unclosed barrier fails compilation.

Use barriers when two head nodes have a real ordering relationship which ordinary per-element classification cannot express. See [Head processing](../sklair/head-processing.md) for the full priority order.
