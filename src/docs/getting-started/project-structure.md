{
  "description": "Understand the source, component, layout, hook, generated, cache, and output directories in a Sklair project."
}
---

# Project structure

Sklair does not impose one sacred directory layout. It resolves configured paths relative to the directory containing `sklair.json`, which means an existing site can keep its shape instead of being reorganised to please the compiler.

The default configuration produced by `sklair init` describes this structure:

```text
project/
├── sklair.json
├── src/
│   └── index.html
├── components/
├── layouts/
├── hooks/
│   ├── pre/
│   └── post/
└── build/
```

## Source and output

`input` is the tree Sklair discovers. HTML-family files are compiled; neighbouring files are copied without interpretation. Their paths relative to `input` become their paths beneath `output`.

`output` is compiler-owned. A successful build replaces it, so do not put authored files there and hope sentimentality will protect them.

## Components, layouts, and hooks

`components` contains reusable component definitions. A component may be one HTML file or a directory with an `index.html` entry and companion assets.

`layouts` contains full HTML documents used by generated documents. Layouts are discovered separately from source pages.

When hooks are enabled, `hooks/pre` and `hooks/post` contain alphabetically ordered Lua files. The configured component, layout, hook, and output directories are automatically excluded from ordinary source discovery when they sit beneath `input`.

## Sklair's workspace

The compiler maintains `.sklair` beside the configuration:

```text
.sklair/
├── cache/
├── generated/
└── temp/
```

Pre-build work starts with `generated` and `temp` cleared. `cache` persists until `sklair clean`, which makes it suitable for hook-owned data which is expensive to reconstruct. Generated assets are copied into `/_sklair/generated/` in the built site.
