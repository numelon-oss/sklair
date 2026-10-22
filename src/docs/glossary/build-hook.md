{
  "description": "Lua run before or after compilation to prepare inputs or inspect and modify build output.",
  "definition": "A build hook is a Lua file which runs before or after Sklair compiles the project."
}
---

# Build hook

A build hook is a Lua file run by Sklair. Pre-build hooks can read project data and create generated inputs; post-build hooks can inspect or modify the completed output.

See [Build hooks](../hooks/index.md).
