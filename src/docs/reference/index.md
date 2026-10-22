{
  "description": "Look up Sklair's commands, project configuration, browser runtime API, and Lua library APIs."
}
---

# Reference

This section describes Sklair as it exists, including the slightly awkward edges which are still awaiting implementation.

## Project surface

- [Command-line interface](./cli.md) covers every command and global logging flag.
- [Project configuration](./project-configuration.md) lists every `sklair.json` field, its default, and whether it currently has an effect.
- [OpenGraph component](./opengraph.md) generates page descriptions and social-sharing metadata without a component file.
- [Compiler directives](../compiler-directives/index.md) defines the comment syntax for head ordering and source removal.

## Lua libraries

- [Markdown](./markdown.md) parses one immutable document for HTML and heading inspection.
- [JSON](./json.md) preserves explicit null, array, and object identities across Lua.
- [JSON Schema](./json-schema.md) compiles Draft 2020-12 schemas and validates JSON or structured Lua values.

## Browser API

- [Sklair Runtime](./runtime-module.md) documents the optional `renderTemplate` module emitted only for documents which register runtime templates.

Everything not explicitly described as browser-time work happens during the build. The reference therefore mostly describes inputs to a compiler whose final product remains ordinary HTML, CSS, and JavaScript.
