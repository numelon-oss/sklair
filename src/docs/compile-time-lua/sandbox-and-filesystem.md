{
  "description": "Understand the libraries and read-only filesystem namespaces available to inline compile-time Lua."
}
---

# Sandbox and filesystem

Lua is safely sandboxed in Sklair to prevent malicious use. The base, table, string, math, and safe portions of the OS library (e.g. `#!lua os.date` and `#!lua os.time`) are available. Other things like process execution, arbitrary file loading and other host operations are not available for obvious reasons.

Lua's `#!lua require` currently reports that module loading is unavailable. There are currently no plans for it to be functional in the traditional sense, given that all Lua files are executed in the hooks directory; it is not a workspace. However, it may be argued that it could be useful in inline Lua.

Shared first-party functionality lives in the provided `json`, `markdown`, and `schema` libraries instead. `#!lua print(...)` writes to Sklair's process output for diagnostics; it does not emit document content.

## How does it work?

Sklair creates one Lua state for a build and executes each use in a fresh child thread ([don't be fooled by "thread"](https://richy.lol/blog/2025/12/29/hitting-the-inherent-concurrency-ceiling-of-lua/), it is actually a [coroutine](https://www.lua.org/pil/9.html)). This gives module and runtime infrastructure one home without letting one block terminate or contaminate every other use.

Scopes decide where deliberate state is shared:

- static blocks in one definition share a definition scope;
- dynamic blocks in one concrete instance share an instance scope;
- every hook receives its own scope;
- different definitions and instances remain isolated.

Within a scope, ordinary Lua globals survive from one block to the next. Threads still make execution boundaries explicit, particularly for `os.exit` and error reporting.

## Read-only filesystem

Inline Lua receives `#!lua fs.read(path)` and `#!lua fs.scandir(path)`, but not `#!lua fs.write`. Build inputs should be prepared in pre-build hooks, before document compilation begins.

Available namespaces are:

| Namespace | Inline access | Meaning |
| --- | --- | --- |
| `project:` | read | configured input tree |
| `cache:` | read | persistent project build cache |
| `temp:` | read | build-local temporary workspace |
| `generated:` | read | assets prepared by pre-build hooks |
| `built:` | unavailable | completed output, which does not exist at the time of inline usage |

Namespace resolution rejects traversal beyond its permitted root. The one intentional project exception is `project:../`, which exposes the directory immediately containing the configured input tree; it exists for project-owned files placed beside `src`, not as a route through the rest of the machine.

Filesystem operations return Lua errors with the resolved namespace context. Use hooks when writes are genuinely part of the build; inline code should only observe the plan it is compiling.
