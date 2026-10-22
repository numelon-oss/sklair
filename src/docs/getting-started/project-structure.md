{
  "description": "Create the source, component, layout, and output directories used by a Sklair project."
}
---

# Project structure

Create an empty directory for the project, enter it, and run:

```bash title="Terminal"
sklair init
```

`sklair init` asks where your source, components, layouts, hooks, and built site should live. Press Enter to accept a default. It writes `sklair.json`; it does not create the directories themselves.

For this guide, use the defaults and create the following structure:

```text title="Project files" hl_lines="3 5"
my-site/
├── sklair.json
├── src/
│   └── index.html
├── components/
│   └── SiteGreeting.html
├── layouts/
└── build/
```

The files can remain empty for now. The next page fills them in.

## What each directory does

| Path | Purpose |
| --- | --- |
| `#!json sklair.json` | Configures the project. Paths inside it are relative to this file. |
| `#!text src/` | Contains pages and static files. HTML is compiled; neighbouring files are copied. |
| `#!text components/` | Contains reusable HTML components. |
| `#!text layouts/` | Contains full-document layouts used by generated documents. It may be empty. |
| `#!text build/` | Receives the completed site. Sklair replaces it after a successful build. |

Sklair does not require this exact layout. Change the paths in `sklair.json` if an existing project is arranged differently.

## Sklair's workspace

During a build, Sklair creates `.sklair` beside the configuration file:

```text title="Compiler-owned workspace"
.sklair/
├── cache/
├── generated/
└── temp/
```

Hooks may keep reusable data in `cache`. Generated inputs and temporary files are cleared when appropriate. `sklair clean` removes the output directory and the entire `.sklair` workspace.
