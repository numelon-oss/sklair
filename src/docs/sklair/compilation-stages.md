{
  "description": "Follow a build through discovery, hooks, output planning, component preparation, compilation, and output."
}
---

# Compilation stages

A Sklair build discovers project files, runs hooks, plans output paths, compiles documents, and writes the completed site.

## 1. Discovery

Sklair resolves paths relative to `sklair.json`, applies default and configured exclusions, and indexes source documents, static files, components, layouts, runtime-template names, and hooks.

Components and layouts are indexed by name. Their HTML is read and parsed only when they are used, so an unused component does no preparation work.

## 2. Pre-build hooks

The temporary and generated workspaces are cleared, one Lua runtime is created for the build, and pre-build hooks execute alphabetically. Hooks can prepare data or assets and call `documents.render` to register generated documents.

Registration does not write HTML. It adds a request to the same document plan used by source files.

## 3. Output planning

Source documents, static files, and generated requests claim their final paths. Generated requests are normalised and sorted. Traversal and collisions fail here, before Sklair removes the previous output. Each output path must have one owner.

## 4. Definition preparation

Source and generated bodies are parsed. When a component or layout is first needed, Sklair parses its HTML, validates it, and runs its static Lua. That prepared source is called a definition and is cached for the rest of the build.

Every use copies the prepared definition to make an instance with its own props and supplied body. For example, two `Badge` components with different labels share the parsed definition but produce different output:

```html
<Badge label="Stable"></Badge>
<Badge label="Experimental"></Badge>
```

Static Lua runs once per used definition per build. Dynamic Lua runs for each instance. See [Static Lua](../lua/static-lua.md) and [Dynamic Lua](../lua/dynamic-lua.md) for their scopes and available values.

## 5. Instantiation and resolution

Bindings apply supplied props, dynamic Lua executes for the particular instance, component bodies are projected into slots, and recursive dependencies are resolved depth-first. Head contributions, runtime-template requirements, and used component folders propagate with the resolved graph.

Generated document bodies are compiled normally, then projected into an instantiated full-document layout.

## 6. Finalisation

Sklair registers native runtime templates when requested, [processes the complete head](./head-processing.md), applies FOUC support when configured, and adds generator metadata. Development-server builds also receive their reload client.

## 7. Replacement and post-build work

The output directory is replaced, compiled documents and component assets are written, static inputs are copied, and only then do post-build hooks run. Post hooks can inspect and alter `built:` because the current output finally exists.

This order is deliberate. A pre-hook can decide what should exist; a post-hook can react to what was successfully built.
