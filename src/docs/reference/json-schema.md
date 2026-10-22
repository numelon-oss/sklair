{
  "description": "Compile Draft 2020-12 JSON Schemas and validate JSON source or structured Lua values."
}
---

# JSON Schema library

The `schema` library compiles JSON Schema Draft 2020-12 and validates either JSON source or an already structured Lua value.

```lua
local validator, err = schema.compile(schema_source, {
  source = "project:schemas/article.schema.json",
})
```

The schema may be a JSON string or a structured Lua object. `source` is an optional path identifying the schema's origin. It is used to resolve relative references and improve diagnostics, so it affects reference resolution as well as error messages. Other declared JSON Schema drafts are rejected; Sklair implements Draft 2020-12 and embeds its official meta-schemas.

Remote references are unavailable. If you really need them, just use the HTTP library to retrieve it and pass that in as if it were a local file.

Relative references remain inside the permitted filesystem namespace and sandbox boundary. Standard format assertions are enabled.

## Validation

```lua
local from_json, err = validator:validate(json_source)
local from_value, err = validator:validate_value(value)
```

`validate` parses a JSON string before validation. `validate_value` accepts Sklair's structured Lua representation directly, which avoids an unnecessary encode/decode round trip when data already came from `json.decode` or was assembled in Lua.

A successful call returns a result object:

```lua
{
  valid = false,
  value = validated_value,
  errors = {
    {
      instancePath = "/title",
      schemaPath = "/required",
      message = "...",
    },
  },
}
```

`valid` is the overall outcome, `value` is the structured value which was checked, and `errors` is an ordered array of failures. Compilation or input-decoding problems use the ordinary `nil, error` return instead of pretending to be schema mismatches.
