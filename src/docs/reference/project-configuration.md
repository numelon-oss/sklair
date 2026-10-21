{
  "description": "Reference every implemented field in sklair.json, its default, and its effect on discovery and compilation."
}
---

# Project configuration

Sklair reads `sklair.json` from the current directory, falling back to `src/sklair.json`. All configured paths are resolved relative to the directory containing that file.

```json
{
  "$schema": "https://github.com/numelon-oss/sklair/releases/latest/download/sklair.schema.json",
  "input": ".",
  "components": "components",
  "layouts": "layouts",
  "hooks": {
    "enabled": true,
    "path": "hooks"
  },
  "output": "../build",
  "templates": ["AudioTier"]
}
```

## Fields

<ProjectConfigurationReference></ProjectConfigurationReference>

HTML extensions `.htm`, `.html`, `.shtml`, and `.xhtml` are compiled. Other discovered files are copied. `excludeCompile` reclassifies matching HTML as a static file; `exclude` removes a match entirely.

Patterns follow gitignore-style doublestar matching, including `!` re-inclusion, leading `/` anchoring, and directory matching. Sklair also excludes its configuration file, `.sklair`, Git and common editor metadata, environment files, `node_modules`, macOS metadata, and configured component, layout, hook, and output directories.

## `#!json hooks`

```json
{
  "hooks": {
    "enabled": true,
    "path": "hooks"
  }
}
```

`path` defaults to `hooks`; Sklair discovers `.lua` files beneath its `pre` and `post` directories. The nested `hooks.http` shape is present in the configuration schema for future HTTP support, but no HTTP module is currently exposed to hooks. Setting those fields does not enable requests.

## `#!json preventFOUC`

```json
{
  "preventFOUC": {
    "enabled": true,
    "colour": "#f5f2ed"
  }
}
```

When enabled, Sklair inserts an early style which hides the document against the configured colour and a body script which reveals it on window load, aiming to prevent a Flash Of Unstyled Content (FOUC). The default colour is `#202020`.
