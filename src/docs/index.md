{
  "description": "Find Sklair's getting-started guide and documentation for components, Lua, hooks, and document generation."
}
---

# Sklair documentation

Sklair is a HTML compiler. It gives ordinary HTML build-time components, explicit logic, and serious document generation, then leaves the browser with documents it already understands.

## Start somewhere useful

If this is your first encounter with Sklair, follow [Getting started](/docs/getting-started/). It takes you from a downloaded binary to a small project with a real component and a development server.

Read [What is Sklair?](./sklair/what-is-sklair.md) for an introduction to its capabilities, or [Compilation stages](./sklair/compilation-stages.md) and [Head processing](./sklair/head-processing.md) for how a build works.

Continue into [Components](/docs/components/), [Lua](/docs/lua/), [Build hooks](/docs/hooks/), [Compiler directives](/docs/compiler-directives/), and [Document generation](/docs/document-generation/) for their usage and APIs. This documentation site is an example of Sklair's powerful document generation.

## What the compiler currently provides

- Recursive HTML components with props, conditional bindings, head contributions, bodies, and colocated assets.
- Static and instance-level Lua which executes during compilation.
- Sandboxed pre- and post-build hooks.
- Full-document layouts and any-number-of-documents generation.
- Markdown to HTML document generation, and structured heading extraction.
- JSON values which preserve arrays, objects, and null.
- JSON Schema Draft 2020-12 validation.
- An optional, deliberately tiny 3KB browser runtime for native templates.
- Deterministic output planning, head ordering, deduplication, and collision errors.

See also: [future work](./future-plans.md), which may or may not be added to Sklair.
