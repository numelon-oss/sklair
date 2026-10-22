{
  "description": "Use sklair.text, sklair.html, and sklair.component to create content from inline Lua."
}
---

# Emitting markup from Lua

Compile-time Lua emits through three deliberately small functions.

## Escaped text

```lua
sklair.text(value)
```

`value` may be a string, number, or boolean. It becomes a text node, so HTML-significant characters are escaped when the document is serialised.

## Trusted HTML

```lua
sklair.html('<strong class="status">Ready</strong>')
```

The string is parsed as a HTML fragment in the script element's exact parent context.

!!! critical "Trusted HTML"

    `#!lua sklair.html` inserts HTML without escaping it. Only pass source you control or have deliberately sanitised. Use `#!lua sklair.text` for plain text from users or other untrusted sources.

## Component invocations

```lua
sklair.component("AudioTier", {
  tier = "master",
  enabled = true,
  redundant = false,
})
```

This emits an unresolved Sklair component invocation. See [the section below](#props) to see how prop values are used.

Sklair sorts emitted attributes so equivalent input remains deterministic.

The named component must exist. Native HTML element names are rejected, and props cannot be supplied to runtime-template registration invocations.

Each Lua block writes into a private buffer. Sklair commits the nodes only after successful execution; an error or non-zero `os.exit` discards the partial result and fails the build.

<!-- TODO: does this comment about recursive lua even belong in the component invocations section?? -->
Emitted markup may invoke components, but it may not emit another compile-time Lua script. For obvious reasons, I hope (It would become a recursive loop).

## Props

Dynamic component and layout scopes expose `props` as a read-only Lua table. Assigning or deleting a prop fails instead of silently changing a compiler-owned snapshot. Top-level prop names are normalised case-insensitively.

HTML attributes convert as follows:

| Source value | Lua value |
| --- | --- |
| missing prop | `#!lua nil` |
| bare or empty attribute | `#!lua true` |
| `"true"` | `#!lua true` |
| `"false"` | `#!lua false` |
| any other attribute string | string |

Generated layouts may receive structured values, including nested arrays, objects, numbers, booleans, strings, and explicit JSON-style nulls. Sklair snapshots them when `documents.render` is called, so mutating the request table later cannot rewrite queued work.
