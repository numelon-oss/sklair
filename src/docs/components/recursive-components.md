{
  "description": "Use components inside other components and understand how their heads, assets, and templates are included."
}
---

# Recursive components

Components may invoke other components exactly as source documents do:

```html
<article class="album-card">
  <AlbumArtwork></AlbumArtwork>
  <AlbumSummary></AlbumSummary>
  <AudioTier></AudioTier>
</article>
```

Sklair resolves the complete graph depth-first in source order. Nested ordinary components expand in place before traversal continues with the owner's remaining nodes.

## Transitive contributions

Every resolved dependency can contribute head nodes, component-folder assets, dynamic behaviour, and runtime-template requirements. Those contributions propagate to the containing document even when the dependency is several components deep.

The same resolved instance contributes its head and folder assets at most once per document. Runtime template names are deduplicated when they arrive through several dependency paths.

## Bodies and props

Bindings run before recursive discovery, so props can determine which dependency branches exist. Caller bodies are compiled in caller scope before being projected, then any components in the definition and projected content continue through ordinary resolution.

Runtime-template dependency trees have one additional restriction: no component in that runtime tree may contain dynamic compile-time Lua. A native template must not secretly depend on per-instance compiler state it can never receive in the browser.
