{
  "description": "Compile-time Lua calculated once for a used definition and reused everywhere that definition appears.",
  "definition": "Static Lua runs once for a used component, layout, or document definition, then reuses its result."
}
---

# Static Lua

Static Lua is calculated once when Sklair first needs a component, layout, or source document. Every later use receives the same prepared result, so static Lua cannot depend on per-use props.

See [Static Lua](/docs/compile-time-lua/static-lua/).
