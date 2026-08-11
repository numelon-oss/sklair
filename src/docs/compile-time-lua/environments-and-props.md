{
  "description": "Understand Lua scopes, threads, shared block state, immutable props, and isolation between definitions and instances."
}
---

# Lua environments and props

Sklair creates one Lua state for a build and executes each use in a fresh child thread ([don't be fooled by "thread"](https://richy.lol/blog/2025/12/29/hitting-the-inherent-concurrency-ceiling-of-lua/), it is actually a [coroutine](https://www.lua.org/pil/9.html)). This gives module and runtime infrastructure one home without letting one block terminate or contaminate every other use.

Scopes decide where deliberate state is shared:

- static blocks in one definition share a definition scope;
- dynamic blocks in one concrete instance share an instance scope;
- every hook receives its own scope;
- different definitions and instances remain isolated.

Within a scope, ordinary Lua globals survive from one block to the next. Threads still make execution boundaries explicit, particularly for `os.exit` and error reporting.

## Props

Dynamic component and layout scopes expose `props` as a read-only Lua table. Assigning or deleting a prop fails instead of silently changing a compiler-owned snapshot. Top-level prop names are normalised case-insensitively.

HTML attributes convert as follows:

| Source value | Lua value |
| --- | --- |
| missing prop | `#!lua nil` |
| bare or empty attribute | `#!lua true` |
| `"true"` | `#!lua true` |
| `"false"` | `#!lua false` |
| any other attribute string | string |

Generated layouts may receive structured values, including nested arrays, objects, numbers, booleans, strings, and explicit JSON-style nulls. Sklair snapshots them when `documents.render` is called, so mutating the request table later cannot rewrite queued work.

`#!lua os.exit(0)` ends the current use successfully. A non-zero code fails that use and therefore the build. It terminates the current thread, not the entire shared Lua state.
