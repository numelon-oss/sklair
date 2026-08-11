{
  "description": "Pass compile-time props through component attributes and consume them with Sklair's explicit binding syntax."
}
---

# Component props and bindings

Attributes on an ordinary component invocation are compile-time props:

```html
<NavigationBar links="false" theme="dark"></NavigationBar>
```

The definition implicitly declares accepted props by referring to them through `sklair-*` bindings or dynamic Lua. There is no separate prop schema to keep synchronised with the markup.

## Conditional nodes

`sklair-if` retains an element when its boolean prop is true and removes the complete subtree otherwise:

```html
<nav>
  <div sklair-if="links">
    <a href="/">Home</a>
  </div>
</nav>
```

An absent boolean is false. A bare attribute, an empty value, or the exact string `true` is true; the exact string `false` is false. Other strings are invalid for boolean bindings.

Bindings execute before recursive component discovery. A component beneath a false condition contributes no body, head nodes, folder assets, Lua execution, or runtime-template requirement.

## Text, attributes, and classes

```html
<span sklair-attr-data-tier="tier" sklair-class-featured="featured">
  <span sklair-text="label"></span>
</span>
```

- `sklair-text="label"` replaces all child nodes with one escaped text node.
- `sklair-attr-data-tier="tier"` writes the `data-tier` attribute.
- `sklair-class-featured="featured"` toggles the `featured` class.

Text and attribute props are required when their binding is reached. Class bindings use boolean semantics. A prop cannot be used as both a string and a boolean in the same definition.

An attribute binding can pass a value into a nested component:

```html
<AudioTier sklair-attr-tier="tier" sklair-attr-label="label"></AudioTier>
```

## Validation and output

Unknown props are rejected unless dynamic Lua opened the definition's prop set. Duplicate attributes, invalid prop names, missing required values, and incompatible binding kinds also fail compilation.

Every compile-time binding is consumed. The final HTML contains its result, never the `sklair-*` instruction which produced it.

`sklair-prop-*` is intentionally absent here. DOM property assignment belongs to runtime templates; ordinary compile-time HTML has attributes and nodes rather than live element properties.
