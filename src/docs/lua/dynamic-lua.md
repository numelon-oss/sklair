{
  "description": "Run Lua for each use of a component or layout, read its props, and emit content based on their values."
}
---

# Dynamic Lua

Dynamic Lua runs once for every concrete document, component, or layout instance. It receives that instance's immutable props and may emit different markup each time.

```html
<script type="application/x-sklair-lua">
  if props.featured then
    sklair.component("Icon", { name = "star" })
    sklair.text("Featured")
  end
</script>
```

Sklair analyses the Lua syntax tree before execution. Literal reads such as `props.name` and `props["name"]` become declared prop requirements, just like a binding. A dynamic key such as `props[key]`, iterating the whole table, or otherwise treating `props` as an open object permits arbitrary supplied props instead.

This catches misspellings without forcing authors to repeat prop declarations in a second schema. Note that this is simple syntax analysis; don't try to do annoying funky stuff.

<!-- TODO: add example, like this: <some code of lua inside a conditional that wont run for example> --->
Bindings and `sklair-if` are resolved before a surviving dynamic block executes. A block inside a false conditional therefore does not run. Markup emitted by Lua is bound again and recursively compiled, so it may contain ordinary component invocations and directives.

Source documents receive an empty props table. Components and layouts receive the props supplied to that instance. Bare attributes, empty values, and the exact string `"true"` become Lua `true`; the exact string `"false"` becomes Lua `false`; other attribute values remain strings. Structured layout props preserve their arrays, objects, scalars, and explicit null values.
