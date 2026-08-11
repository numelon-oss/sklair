{
  "description": "Install Sklair, create a project, and compile your first site into ordinary HTML."
}
---

# Getting started

This guide takes the shortest honest route from a Sklair binary to a compiled site. You will create a project, define one component, use it from a HTML document, and run the development server.

Sklair does not require Node.js, a package manifest, or a browser runtime. The compiler is a standalone Go binary and the project can begin with little more than `sklair.json`, a HTML file, and a components directory.

## The route through this guide

1. [Install Sklair](/download) and confirm the binary runs.
2. Learn the [project structure](/docs/getting-started/project-structure/) produced by `sklair init`.
3. Compile [your first build](/docs/getting-started/your-first-build/) and inspect the output.
4. Use the [development server](/docs/getting-started/development-server/) while changing the project.

You can read those pages in order, although there is nothing sacred about the sequence. Existing projects can add Sklair by writing a configuration file directly and keeping their current directory structure.

## What you should expect

The first example is intended to be small, demonstrating the important basics of Sklair: a component invocation is present in source and absent from output.
Once that contract is clear, recursive components, Lua, runtime templates, and document generation are extensions of the same compilation model.

Sklair replaces the configured output directory on each successful build. Keep authored source outside that directory and treat built files as disposable artefacts.
