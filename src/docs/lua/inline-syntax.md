{
  "description": "Embed Lua in HTML with application/x-sklair-lua script tags and select static or dynamic execution."
}
---

# Inline Lua syntax

Embed Lua in a script element with Sklair's MIME type:

```html
<script type="application/x-sklair-lua">
  sklair.text("This is escaped text.")
</script>
```

The MIME type says exactly what is happening: Sklair, rather than the browser, owns this script. The element is removed during compilation and replaced by whatever the block emits.

HTML already treats script contents as raw text, so comparisons and operators containing `<` or `>` need no custom parser or escaping ritual. The normal HTML closing rule still applies: a literal `</script>` ends the element, even inside a Lua string.

Add the boolean `static` attribute for definition-level execution:

```html
<script type="application/x-sklair-lua" static>
  sklair.component("Icon", { name = "sparkles" })
</script>
```

Inline code must live inside the element. `src` is rejected because external compile-time execution would make ownership and discovery needlessly ambiguous. Duplicate `type` attributes are rejected as malformed source.

Compile-time Lua can appear in source documents, ordinary components, generated bodies, and layouts. Runtime-template definitions cannot contain it: those definitions describe browser-time clones, so running part of one at compile time would give the same source two incompatible meanings.
