{
  "description": "Use embedded Lua during compilation without turning Lua into a browser runtime or general template language."
}
---

# Compile-time Lua

Compile-time Lua is an escape hatch inside Sklair's compiler which runs while HTML is being built, emits ordinary nodes, and disappears from the final document.

```html
<script type="application/x-sklair-lua">
  sklair.html("<p>Built at " .. os.date("%H:%M") .. "</p>")
</script>
```

Use it when HTML structure genuinely depends on data or computation: generating a navigation tree, rendering a Markdown-derived table of contents, or adapting one component to structured input. Props and bindings should handle the simple cases. Lua handles the cases where pretending HTML is an expression language would be worse.

Sklair has two execution modes:

- dynamic Lua runs for each document, component, or layout instance;
- static Lua runs once while a used definition is prepared, and its unresolved output is reused (effectively cached).

Every block executes in its own isolated and [sandboxed](./sandbox-and-filesystem.md) Lua thread.

`#!lua os.exit(0)` ends the current use successfully. A non-zero code fails that use and therefore the build. It terminates the current thread, not the entire shared Lua state.
