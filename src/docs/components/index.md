{
  "description": "Create reusable HTML components with props, bodies, head contributions, and colocated assets."
}
---

# Components

Sklair components are reusable HTML definitions expanded during compilation. They can contribute markup where invoked, nodes to the document head, owned assets, and requirements for the optional runtime template registry.

An invocation is an ordinary non-standard element whose name matches a discovered component:

```html title="Component example"
<ReleaseCard featured="true"></ReleaseCard>
```

After compilation, that tag is gone. Its resolved body occupies the same position and its head contribution has joined the final document head.

## Components remain HTML

A definition is parsed as a HTML document. Its `<head>` describes document-level dependencies and its `<body>` contains the markup which replaces an invocation. There is no second component language and no string interpolation syntax hidden between the tags.

Use [props and bindings](./props-and-bindings.md) when an instance needs scalar variation. Use a [body and default slot](./bodies-and-slots.md) when the caller owns content inside the reusable shell. Use [compile-time Lua](../lua/index.md) when the decision is genuinely more involved than a binding.

## Composition

Definitions may invoke other components. Sklair resolves that graph recursively, propagates head contributions and assets, and reports cycles with their dependency chain. A component can therefore stay small without making the final document pay for a runtime component manager.

## Two physical forms

A simple component is one HTML file. A component which owns CSS, modules, images, or fonts can become a [component folder](./folders-and-assets.md) with `index.html` as its entry point. Relative resource URLs are then rebased into the component's built namespace.

## Two execution modes

Ordinary components compile immediately. Components listed in the root `templates` configuration become [runtime templates](./runtime-templates.md): their body is registered as native browser template content and cloned later through Sklair's small ES module.

The source syntax is shared, but the semantics are deliberately distinct. Runtime-template registration cannot receive compile-time props or a compile-time body, and runtime template definitions cannot contain compile-time Lua.
