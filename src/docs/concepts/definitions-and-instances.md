{
  "description": "Understand the difference between reusable prepared definitions and the component, layout, and document instances compiled from them."
}
---

# Definitions and instances

A definition is reusable prepared source. An instance is one concrete use of that definition with a particular prop set and, for ordinary components, a particular projected body.

This distinction lets Sklair avoid reparsing the same component while still efficiently producing different markup for each invocation.

## Component definitions

When a component is first needed, Sklair reads and parses its entry HTML, separates head and body nodes, validates its slot and runtime mode, and executes its static Lua. That prepared definition is cached for the rest of the build.

An unused component is never prepared. Static does not mean "execute every file in the directory to satisfy my perception of completeness".

## Component instances

For an ordinary invocation, Sklair clones the definition, binds attributes as props, projects caller content, executes dynamic Lua, and resolves nested components. Prop values form part of the instance identity used while resolving the graph.

```html
<Badge label="Stable"></Badge>
<Badge label="Experimental"></Badge>
```

Those invocations share one parsed definition and produce two instances.

## Layout and document instances

A full-document layout is also prepared once and instantiated for each generated request. Unlike component attributes, its props may contain nested arrays and objects because they enter through `documents.render`.

A generated body is a document definition in its own right. It receives empty document props and compiles before being projected into the layout. Layout props belong to the layout; they are not implicitly smuggled into the body.

## Static and dynamic behaviour

Static inline Lua belongs to definition preparation and can be cached safely. Dynamic Lua belongs to an instance and can observe its immutable props. Confusing the two would either repeat expensive work unnecessarily or, worse, reuse one caller's values for another caller. Sklair makes the lifetime explicit because caches should follow semantics, not optimism.
