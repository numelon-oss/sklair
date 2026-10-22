{
  "description": "Decode and encode JSON in Lua and preserve null values, arrays, and objects."
}
---

# JSON library

Sklair's `json` library converts between JSON and structured Lua values while preserving the distinctions Lua normally loses around nulls and empty containers.

## Decode and encode

```lua
local value, err = json.decode(source)
local source, err = json.encode(value)
```

Each function accepts exactly one argument and returns either the result or `nil, error`. Decoded objects become keyed tables, arrays become indexed tables with preserved array identity, and JSON null becomes `json.null`.

Supported encoded values are strings, finite numbers, booleans, `json.null`, and valid nested tables. Cycles, sparse arrays, mixed numeric and object keys, unsupported userdata, and non-finite numbers are errors.

## Arrays and Objects

An empty Lua table is ambiguous, so unmarked `{}` encodes as an object. Use constructors when identity must be explicit:

```lua
local empty_array = json.array()
local empty_object = json.object()

local tags = json.array({ "compiler", "html" })
local person = json.object({ name = "Nu Melonseed" })
```

Both constructors optionally accept one table and preserve its array or object identity for later JSON encoding and structured compiler APIs. Non-empty, unmarked contiguous numeric tables are inferred as arrays; keyed tables are inferred as objects.

The marker belongs to the table value, so ordinary mutations retain the intended container kind as long as the resulting structure remains valid for that kind. `json.null` is an explicit singleton value and is distinct from Lua `nil`, which means absence.
