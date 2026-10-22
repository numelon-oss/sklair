{
  "description": "Learn what Sklair is, why it exists, and how it extends HTML without replacing the web platform."
}
---

# Sklair documentation

Sklair is a HTML compiler. It gives ordinary HTML build-time components, explicit logic, and serious document generation, then leaves the browser with documents it already understands.

That distinction is the whole project. Sklair is *not* a framework which happens to export HTML after doing something considerably stranger. Your source begins as HTML, the compiler extends it in specific areas where traditional HTML may have fallen short, and the result is still HTML.

## Start somewhere useful

If this is your first encounter with Sklair, follow [Getting started](/docs/getting-started/). It takes you from a downloaded binary to a small project with a real component and a development server.

If the syntax already makes sense but the design does not, read [Concepts](/docs/concepts/). The compilation model, the distinction between definitions and instances, and Sklair's insistence on bare output explain most of the decisions elsewhere.

When you are building something less polite than a five-page website, continue into [Components](/docs/components/), [Compile-time Lua](/docs/compile-time-lua/), and [Document generation](/docs/document-generation/). Those facilities are how the same compiler has ended up underneath large websites and full browser applications without acquiring a virtual DOM out of panic. This documentation site is an example of Sklair's powerful document generation.

## What the compiler currently provides

- Recursive HTML components with props, conditional bindings, head contributions, bodies, and colocated assets.
- Static and instance-level Lua which executes during compilation.
- Sandboxed pre- and post-build hooks.
- Full-document layouts and any-number-of-documents generation.
- Markdown to HTML document gneration, and structured heading extraction.
- JSON values which preserve arrays, objects, and null.
- JSON Schema Draft 2020-12 validation.
- An optional, deliberately tiny 3KB browser runtime for native templates.
- Deterministic output planning, head ordering, deduplication, and collision errors.

See also: [future work](./future-plans.md), which may or may not be added to Sklair.
