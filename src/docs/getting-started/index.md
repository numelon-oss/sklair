{
  "description": "Create a small Sklair project, compile a component into ordinary HTML, and use the development server."
}
---

# Getting started

This guide creates a small Sklair project and compiles its first [component](../glossary/component.md). You only need a Sklair binary and a terminal.

Download Sklair, mark it as executable if your platform requires it, and place it on `PATH` using the [download page](/download/). Then confirm it runs:

```bash title="Terminal"
sklair version
```

## Build the project

Follow these pages in order:

1. Set up the [project structure](./project-structure.md).
2. Compile [your first build](./your-first-build.md).
3. Work with the [development server](./development-server.md).

The example uses one source document and one component. By the end, the component tag and its Sklair attributes will have disappeared into ordinary HTML.

Sklair replaces the configured output directory after every successful build. Keep your source outside it and treat built files as disposable.
