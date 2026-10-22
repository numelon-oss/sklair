{
  "description": "Run Lua inside HTML or build hooks, with shared libraries, isolated scopes, and filesystem access."
}
---

# Lua

Lua runs during the build. You can embed it inside HTML or write [build hooks](../hooks/index.md) which run before or after compilation.

## Inline Lua

Inline Lua runs while HTML is being built, emits ordinary nodes, and disappears from the final document.

```html
<script type="application/x-sklair-lua">
  sklair.html("<p>Built at " .. os.date("%H:%M") .. "</p>")
</script>
```

Use it when HTML structure genuinely depends on data or computation: generating a navigation tree, rendering a Markdown-derived table of contents, or adapting one component to structured input. Props and bindings should handle the simple cases. Lua handles the cases where pretending HTML is an expression language would be worse.

Sklair has two execution modes:

- dynamic Lua runs for each document, component, or layout instance;
- static Lua runs once while a used definition is prepared, and its unresolved output is reused (effectively cached).

See [Inline syntax](./inline-syntax.md), [Static Lua](./static-lua.md), [Dynamic Lua](./dynamic-lua.md), and [Emitting markup](./emitting-markup.md) for usage.

## Shared environment and libraries

[Lua environment](./environment.md) describes the available standard libraries, execution scopes, diagnostics, and `os.exit`. [Filesystem](./filesystem.md) lists the namespaces and permissions for inline Lua and both hook phases.

The [JSON](../reference/json.md), [Markdown](../reference/markdown.md), and [JSON Schema](../reference/json-schema.md) libraries are available in both inline Lua and hooks.
