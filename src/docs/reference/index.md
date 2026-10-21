{
  "description": "Look up Sklair's commands, configuration, directives, runtime API, Lua libraries, and errors."
}
---

# Reference

This section describes Sklair as it exists, including the slightly awkward edges which are still awaiting implementation.

## Project surface

- [Command-line interface](./cli/) covers every command and global logging flag.
- [Project configuration](./project-configuration/) lists every `sklair.json` field, its default, and whether it currently has an effect.
- [Compiler directives](./directives/) defines the comment syntax for head ordering and source removal.

## Lua libraries

- [Markdown](./markdown/) parses one immutable document for HTML and heading inspection.
- [JSON](./json/) preserves explicit null, array, and object identities across Lua.
- [JSON Schema](./json-schema/) compiles Draft 2020-12 schemas and validates JSON or structured Lua values.

## Browser API

- [Sklair Runtime](./runtime-module/) documents the optional `renderTemplate` module emitted only for documents which register runtime templates.

Everything not explicitly described as browser-time work happens during the build. The reference therefore mostly describes inputs to a compiler whose final product remains ordinary HTML, CSS, and JavaScript.
