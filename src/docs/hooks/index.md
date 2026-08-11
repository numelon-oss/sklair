{
  "description": "Run deterministic sandboxed Lua before or after compilation to orchestrate project-specific build work."
}
---

# Build hooks

Hooks let a project orchestrate work around Sklair's compiler without moving the compiler itself into Lua.

Configure a hooks directory, then place `.lua` files in its `pre` and `post` subdirectories:

```text
hooks/
├── pre/
│   └── 10-documents.lua
└── post/
    └── 10-manifest.lua
```

Files execute alphabetically within each phase. Prefixes are a convenient way to make dependencies visible instead of relying on somebody remembering that `build-search.lua` happens to sort after `articles.lua`.

Sklair uses one Lua runtime per build, but every hook has its own isolated scope and child thread (coroutine). Globals do not leak from one hook file into the next. The standard sandbox, `json`, `markdown`, and `schema` libraries are available, while filesystem permissions and compiler modules depend on the phase.

Pre-build hooks prepare inputs, generated assets, and generated-document requests. Post-build hooks inspect or adjust completed output. If logic belongs inside one component instance, use compile-time Lua instead; hooks are for build-level orchestration.
