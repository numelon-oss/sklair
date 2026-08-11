{
  "description": "Learn what Sklair is, why it exists, and how it extends HTML without replacing the web platform."
}
---

# Sklair documentation

Sklair is a HTML compiler. It gives ordinary HTML build-time components, explicit logic, and serious document generation, then leaves the browser with documents it already understands.

That distinction is the whole project. Sklair is *not* a browser framework which happens to export HTML after doing something considerably stranger. Your source begins as HTML, the compiler extends it at deliberate seams, and the result is still HTML.

## Start somewhere useful

If this is your first encounter with Sklair, follow [Getting started](/docs/getting-started/). It takes you from a downloaded binary to a small project with a real component and a development server.

If the syntax already makes sense but the design does not, read [Concepts](/docs/concepts/). The compilation model, the distinction between definitions and instances, and Sklair's insistence on bare output explain most of the decisions elsewhere.

When you are building something less polite than a five-page website, continue into [Components](/docs/components/), [Compile-time Lua](/docs/compile-time-lua/), and [Document generation](/docs/document-generation/). Those facilities are how the same compiler has ended up underneath large websites and full browser applications without acquiring a virtual DOM out of panic. This documentation site is an example of Sklair's powerful document generation.

## What the compiler currently provides

- Recursive HTML components with props, conditional bindings, head contributions, bodies, and colocated assets.
- Static and instance-level Lua which executes during compilation.
- Sandboxed pre- and post-build hooks.
- Full-document layouts and any-number-of-documents generation.
- Immutable Markdown documents and structured heading extraction.
- JSON values which preserve arrays, objects, and null.
- JSON Schema Draft 2020-12 validation.
- An optional, deliberately tiny 3KB browser runtime for native templates.
- Deterministic output planning, head ordering, deduplication, and collision errors.

Package-manager distribution, update notifications, hook HTTP, and reusable package registries are not quietly hiding in here. They remain [future work](./future-plans.md) and the documentation does not pretend otherwise.

## Read by task

Use the [project configuration reference](/docs/reference/project-configuration/) when you need exact `sklair.json` fields, the [CLI reference](/docs/reference/cli/) for commands and flags, and [Errors and diagnostics](/docs/reference/errors/) when the compiler refuses a build. It usually has a reason. Occasionally it even phrases that reason politely.
