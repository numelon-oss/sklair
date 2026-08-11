{
  "description": "Emit escaped text, trusted HTML fragments, and unresolved component invocations from compile-time Lua."
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

<!-- TODO: this section is a good argument for introducing admonitions. something as important as raw html not being trusted should be a big fat warning! not some delightfully composed prose, sleekly hidden. -->

The string is parsed as a HTML fragment in the script element's exact parent context. It is trusted markup, not escaped content; only pass source you control or have deliberately sanitised.

## Component invocations

```lua
sklair.component("AudioTier", {
  tier = "master",
  enabled = true,
  redundant = false,
})
```

This emits an unresolved Sklair component invocation. Prop values may be strings, numbers, booleans, or nil. `true` becomes a bare attribute, while `false` and nil omit the attribute. Sklair sorts emitted attributes so equivalent input remains deterministic.

The named component must exist. Native HTML element names are rejected, and props cannot be supplied to runtime-template registration invocations.

Each Lua block writes into a private buffer. Sklair commits the nodes only after successful execution; an error or non-zero `os.exit` discards the partial result and fails the build.

<!-- TODO: does this comment about recursive lua even belong in the component invocations section?? -->
Emitted markup may invoke components, but it may not emit another compile-time Lua script. For obvious reasons, I hope (It would become a recursive loop).
