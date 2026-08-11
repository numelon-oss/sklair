{
  "description": "Register selected components as native browser templates and render lightweight runtime instances only where needed."
}
---

# Runtime templates

Runtime templates are the narrow exception to Sklair's runtime-free output. They exist for the rare instanc where markup must be created after page load. Good thing it is insanely simple.

List template components at the root of `sklair.json`:

```json
{
  "templates": ["AudioTier"]
}
```

Then use the otherwise-empty component once in any document which needs it:

```html
<AudioTier></AudioTier>
```

That invocation registers the template; it does not render a visible instance. Sklair writes one hidden template registry and `/_sklair/runtime.mjs` only for documents which require them.

## Rendering an instance

```js
import { renderTemplate } from "/_sklair/runtime.mjs"

const badge = renderTemplate("AudioTier", {
  label: "Master",
  tier: "master",
  spec: "24-bit / 96 kHz",
})

target.append(badge)
```

`renderTemplate` returns a fresh `DocumentFragment`. The name is trimmed and matched case-insensitively. An unregistered template is an error.

## Runtime bindings

Template source uses the same spelling as compile-time bindings:

```html
<span sklair-if="available"
      sklair-attr-data-tier="tier"
      sklair-prop-hidden="hidden"
      sklair-class-featured="featured">
  <span sklair-text="label"></span>
</span>
```

Sklair rewrites these to generated `data-sklair-*` attributes inside the native template. Authors must never write that generated namespace themselves.

`sklair-if` removes false subtrees. Text, attribute, property, and class bindings require their named value. Nullish attribute values remove the target attribute; property bindings assign the live DOM property.

Bindings run *once* while cloning - there is **NO REACTIVITY** (yuck) or expression language.

## Runtime bodies

A template may expose one `<slot></slot>`. Pass a `Node`, `DocumentFragment`, or iterable of nodes as the third argument:

```js
renderTemplate("Card", { title: "Album" }, artwork)
```

Omitting the body removes the marker. Supplying a body to a template without a slot is an error. Strings are not treated as HTML.

## Registration rules

Registration invocations cannot have attributes or children and may appear only once explicitly in one document or component source. Requirements reached transitively are deduplicated. Template definitions cannot contain compile-time Lua, their heads cannot contain runtime bindings, and a dependency beneath a template cannot contain dynamic compile-time Lua.
