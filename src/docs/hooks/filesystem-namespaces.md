{
  "description": "Read and write the project, cache, temporary, generated, and built filesystem namespaces from the appropriate hook phase."
}
---

# Hook filesystem namespaces

Hook paths begin with an explicit namespace:

```lua
local source = fs.read("project:content/index.md")
fs.write("generated:search/index.json", encoded_index)
```

The filesystem API consists of `fs.read(path)`, `fs.write(path, content)`, and `fs.scandir(path)`. Directory scans return a deterministically ordered array of entries with `name` and `isDir` fields.

| Namespace | Lifetime | Pre | Post |
| --- | --- | --- | --- |
| `project:` | source-controlled input | read only | read only |
| `cache:` | persists between builds | read/write | read/write |
| `temp:` | cleared at build start | read/write | read/write |
| `generated:` | cleared at build start | read/write staging area | built `/_sklair/` subtree |
| `built:` | current completed output | unavailable | read/write |

`project:` maps to the configured input directory. `project:../` deliberately exposes its immediate parent for project files which live beside that directory; traversal beyond the allowed boundary is rejected.

Use `cache:` only for values which remain valid across builds and can be regenerated. Use `temp:` for intermediates which should never reach output. Use `generated:` for assets that should be published under `/_sklair/`.

The `temp:` and pre-build `generated:` trees are reset at the beginning of every build. `cache:` is not. A clean command removes output, temporary files, and generated staging data; the persistent cache remains intact.
