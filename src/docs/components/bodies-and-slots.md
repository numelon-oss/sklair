{
  "description": "Project caller-owned content into a component through its single default slot."
}
---

# Component bodies and slots

An ordinary component can accept caller-owned content through one default slot:

<!-- TODO: add line highlighting specifically to show the line containing the slot usage -->

```html
<!-- components/Card.html -->
<!doctype html>
<html>
<head></head>
<body>
  <article class="card">
    <slot></slot>
  </article>
</body>
</html>
```

```html
<Card>
  <h2>Numelon Passport</h2>
  <p>Sign in once and continue anywhere.</p>
</Card>
```

The caller's nodes replace `<slot>` during compilation.

## Scope

Caller content compiles in the caller's scope before projection. Components and props inside the body therefore belong to the caller, while bindings and Lua in `Card` belong to the card instance.

This remains true through recursion. A slot may sit inside a nested component invocation and the projected body still arrives at that eventual marker.

## Slot rules

A component body may contain zero or one `<slot></slot>`. The slot must:

- appear in the component body, not anywhere else like the head;
- have no name or other attributes;
- contain no fallback children; and
- be the only slot in the definition.

Supplying children to a component without a slot is an error. Text, comments, and whitespace are real child nodes, so this formatting is not empty:

```html
<Icon>
</Icon>
```

Use `<Icon></Icon>` when the registration or invocation must have no body.

An empty invocation of a slotted ordinary component simply removes the slot. Runtime templates reuse the `<slot>` source marker for a different, browser-time body contract; they cannot receive children on their compile-time registration tag.
