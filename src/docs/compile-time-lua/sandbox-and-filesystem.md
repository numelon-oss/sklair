{
  "description": "Understand the libraries and read-only filesystem namespaces available to inline compile-time Lua."
}
---

# Lua sandbox and filesystem

Inline Lua is capable enough to derive markup, without becoming a malicious backdoor.

The base, table, string, math, and safe portions of the OS library are available. Useful deterministic building blocks such as `#!lua os.date` and `#!lua os.time` therefore remain. Process execution, environment mutation, arbitrary file loading, and similar host-level operations do not.

`#!lua require` currently reports that module loading is unavailable. I do not have any plans for it for now.

Shared first-party functionality lives in the provided `json`, `markdown`, and `schema` libraries instead. `#!lua print(...)` writes to Sklair's process output for diagnostics; it does not emit document content.

## Read-only filesystem

Inline Lua receives `#!lua fs.read(path)` and `#!lua fs.scandir(path)`, but not `#!lua fs.write`. Build inputs should be prepared in pre-build hooks, before document compilation begins.

Available namespaces are:

| Namespace | Inline access | Meaning |
| --- | --- | --- |
| `project:` | read | configured input tree |
| `cache:` | read | persistent project build cache |
| `temp:` | read | build-local temporary workspace |
| `generated:` | read | assets prepared by pre-build hooks |
| `built:` | unavailable | completed output, which does not exist yet |

Namespace resolution rejects traversal beyond its permitted root. The one intentional project exception is `project:../`, which exposes the directory immediately containing the configured input tree; it exists for project-owned files placed beside `src`, not as a route through the rest of the machine.

Filesystem operations return Lua errors with the resolved namespace context. Use hooks when writes are genuinely part of the build; inline code should only observe the plan it is compiling.
