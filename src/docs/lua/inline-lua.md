{
  "description": "Embed Lua in HTML and choose between dynamic execution for each use and static execution once per definition."
}
---

# Inline Lua

Embed Lua in a script element with Sklair's MIME type:

```html
<script type="application/x-sklair-lua">
  sklair.text("This is escaped text.")
</script>
```

The MIME type says exactly what is happening: Sklair, rather than the browser, owns this script. The element is removed during compilation and replaced by whatever the block emits.

HTML already treats script contents as raw text, so comparisons and operators containing `<` or `>` need no custom parser or escaping ritual. The normal HTML closing rule still applies: a literal `</script>` ends the element, even inside a Lua string.

Inline code must live inside the element. `src` is rejected because external compile-time execution would make ownership and discovery needlessly ambiguous. Duplicate `type` attributes are rejected as malformed source.

Compile-time Lua can appear in source documents, ordinary components, generated bodies, and layouts. Runtime-template definitions cannot contain it: those definitions describe browser-time clones, so running part of one at compile time would give the same source two incompatible meanings.

## Dynamic Lua

Without the `static` attribute, inline Lua runs once for every concrete document, component, or layout instance. It receives that instance's immutable props and may emit different markup each time.

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

Bindings and `sklair-if` are resolved before a surviving dynamic block executes. A block inside a false conditional therefore does not run:

```html
<section sklair-if="featured">
  <script type="application/x-sklair-lua">
    sklair.text("This runs only when featured is true.")
  </script>
</section>
```

Markup emitted by Lua is bound again and recursively compiled, so it may contain ordinary component invocations and directives.

Source documents receive an empty props table. Components and layouts receive the props supplied to that instance. Bare attributes, empty values, and the exact string `"true"` become Lua `true`; the exact string `"false"` becomes Lua `false`; other attribute values remain strings. Structured layout props preserve their arrays, objects, scalars, and explicit null values.

## Static Lua

Add the boolean `static` attribute to calculate something once and reuse the result:

```html
<script type="application/x-sklair-lua" static>
  build_label = os.date("%Y-%m-%d")
  sklair.html('<small class="build-date">' .. build_label .. '</small>')
</script>
```

Sklair runs the block when that document, component, or layout definition is first needed, stores its unresolved emitted nodes, and clones those nodes for every instance. If a component is used on twenty pages, its static Lua still runs only once during that build. If the component is unused, its static Lua never runs.

Blocks in one definition execute in source order and share that definition's environment. A later static block can therefore read a global assigned by an earlier one. Head blocks run before body blocks because the definition itself is prepared in document order.

Static Lua cannot read component props or bodies: those values belong to individual instances which do not exist yet. For the same reason, a static block cannot sit beneath `sklair-if`; its execution cannot truthfully depend on an instance-level condition.

The cached output remains unresolved until instantiation, which means emitted component invocations and bindings still receive the correct surrounding context later.

Use static Lua for definition-wide derived markup or setup whose answer is identical across every instance. If props can change the answer, it is dynamic Lua.

See [Emitting markup](./emitting-markup.md) for the output functions and [Lua environment](./environment.md) for shared scopes and exit behaviour.
