{
  "description": "Reference the optional browser runtime's renderTemplate function, values, bodies, return value, and failures."
}
---

# Sklair Runtime module

Documents which register at least one runtime template receive the generated ES module `/_sklair/runtime.mjs`. Documents with no runtime work receive neither the module nor a template registry.

```js
import { renderTemplate } from "/_sklair/runtime.mjs"

const fragment = renderTemplate("AudioTier", {
  label: "Master",
  tier: "master",
  hidden: false,
})
```

## `#!js renderTemplate(name, values = {}, body)`

Clone a registered native template, apply its one-time bindings, optionally project a body, and return a fresh `DocumentFragment`.

Template names are trimmed and matched case-insensitively. `values` is a key-value object. A missing template or missing value required by a binding throws an error.

Binding behaviour is:

| Source binding | Runtime behaviour |
| --- | --- |
| `sklair-if="key"` | Remove the subtree when the value is false. |
| `sklair-text="key"` | Set text content; nullish values become an empty string. |
| `sklair-attr-name="key"` | Set the attribute; nullish values remove it. |
| `sklair-prop-name="key"` | Assign the live DOM property directly. |
| `sklair-class-name="key"` | Toggle the class from the value's boolean truth. |

Sklair rewrites source bindings into an internal `data-sklair-*` representation inside the registry. Do not author that generated namespace directly.

The optional body may be a `Node`, `DocumentFragment`, or iterable of nodes. Nodes are moved into the clone, following ordinary DOM semantics; they are not duplicated. Strings are rejected rather than interpreted as HTML. Omitting the body removes the template's slot marker, while supplying one to a template with no slot throws an error.

Bindings run only while cloning. The returned fragment has **NO** reactive connection to `values`, **NO** component lifecycle, and **NO** framework state. It is simply prepared DOM.
