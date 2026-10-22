{
  "description": "Configure Lua hooks to prepare inputs before compilation or work with completed output afterwards."
}
---

# Build hooks

Hooks let a project orchestrate work around Sklair's compiler without moving the compiler itself into Lua. It is different from [inline Lua](../lua/inline-lua.md), which is for logic inside component instances.

[Pre-build hooks](./pre.md) prepare inputs, generated assets, and generated-document requests. [Post-build hooks](./post.md) inspect or adjust completed output.

Configure a hooks directory, then place `.lua` files in its `pre` and `post` subdirectories:

```text
hooks/
├── pre/
│   └── 10-documents.lua
└── post/
    └── 10-manifest.lua
```

Files execute alphabetically within each phase. Numeric prefixes are a convenient way to show the order in which these hooks run.

See [Lua environment](../lua/environment.md) for execution scopes, standard libraries, diagnostics, and exit handling. Globals do not leak from one hook file into the next. [Filesystem](../lua/filesystem.md) lists the permissions available in each phase.
