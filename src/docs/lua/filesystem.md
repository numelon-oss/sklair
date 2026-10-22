{
  "description": "Read project files and use cache, temporary, generated, and built namespaces from inline Lua and hooks."
}
---

# Filesystem

Lua filesystem paths begin with an explicit namespace:

```lua
local source = fs.read("project:content/index.md")
fs.write("generated:search/index.json", encoded_index)
```

The filesystem API consists of `#!lua fs.read(path)`, `#!lua fs.write(path, content)`, and `#!lua fs.scandir(path)`. Inline Lua inside HTML receives only `read` and `scandir`; writes belong in hooks. Directory scans return a deterministically ordered array of entries with `name` and `isDir` fields.

| Namespace | Lifetime | Inline | Pre-build | Post-build |
| --- | --- | --- | --- | --- |
| `project:` | project inputs | read only | read only | read only |
| `cache:` | persists between builds | read only | read/write | read/write |
| `temp:` | cleared at build start | read only | read/write | read/write |
| `generated:` | staging cleared at build start; published under `/_sklair/` | read staging | read/write staging | read/write built `/_sklair/` subtree |
| `built:` | current completed output | unavailable | unavailable | read/write |

`project:` maps to the configured input directory. `project:../` deliberately exposes its immediate parent for project files which live beside that directory; traversal beyond the allowed boundary is rejected.

Use `cache:` only for values which remain valid across builds and can be regenerated. Use `temp:` for intermediates which should never reach output. Use `generated:` for assets that should be published under `/_sklair/`.

The `temp:` and pre-build `generated:` trees are reset at the beginning of every build. `cache:` is not. `#!bash sklair clean` removes the output directory and the entire `.sklair` workspace, including persistent cache data.
