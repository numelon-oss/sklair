{
  "description": "Understand how Sklair combines head contributions, removes duplicates, and orders metadata, styles, and scripts."
}
---

# Head processing

After resolving components, Sklair processes the complete document head. This includes the source document's own head and every [component head contribution](../components/head-contributions.md).

Sklair groups the head into segments, removes duplicate ordinary elements, and sorts the remaining segments by priority. An ordinary element forms one segment. An [ordering barrier](../compiler-directives/ordering-barriers.md) groups several nodes into one segment whose internal order is preserved.

## Duplicate elements

Sklair keeps the first ordinary element with a matching weak hash and drops later matches. The hash uses the node type, tag name, and attributes in their existing order. For `script` and `style`, it also includes the first text child's contents.

For example, two components contributing this same link produce one stylesheet link:

```html
<link rel="stylesheet" href="/styles/shared.css">
```

This is a lightweight comparison, not a check that two elements have the same browser behaviour. Attribute order affects the hash. Different titles are not compared by their text, so declaring several titles is not a way to override an earlier one. Keep document metadata intentional.

Nodes inside ordering barriers are not individually deduplicated. A link inside a barrier does not suppress an ordinary matching link elsewhere in the head.

## Output order

Segments are ordered as follows:

| Priority | Content |
| --- | --- |
| 1 | `meta` with a `charset` attribute |
| 2 | `meta` with an `http-equiv` attribute |
| 3 | `viewport` and `theme-color` metadata |
| 4 | Resource hints: `preload`, `preconnect`, `prefetch`, `dns-prefetch`, and `modulepreload` links |
| 5 | Sklair's optional FOUC prevention script |
| 6 | Document title |
| 7 | Other `link` elements with a `rel` attribute |
| 8 | Scripts, including the development server's reload client when serving |
| 9 | Segments explicitly marked `analytics` by an ordering barrier |
| 10 | Other elements, including ordinary inline `style` elements |
| 11 | Sklair's generator metadata |

Order within a priority group is not guaranteed. Use an ordering barrier when several elements must keep their order. Non-resource `link` elements are classified in the stylesheet group even when their `rel` is not `stylesheet`. Inline `style` elements are not automatically put in that group; use a barrier with `treat-as=stylesheet` when they need to stay beside stylesheet links.

Sklair orders resource hints you supply. It does not generate preload or preconnect links automatically.

## Comments and removal

Ordinary head comments and whitespace outside barriers are omitted during segmentation. Comments within a barrier are retained with its other nodes, except compiler directive markers.

[Removal regions](../compiler-directives/removal-regions.md) omit their contents before deduplication and sorting. Ordering-barrier and removal marker comments also disappear from the final document.
