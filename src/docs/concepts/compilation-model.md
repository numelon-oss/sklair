{
  "description": "Follow a Sklair build from discovery and hooks through definitions, document compilation, finalisation, and output."
}
---

# Compilation model

A Sklair build is a sequence of ownership transfers. Project files begin as authored input, hooks may register additional work, definitions become concrete documents, and only a fully planned build is allowed to replace the output.

## 1. Discovery

Sklair resolves paths relative to `sklair.json`, applies default and configured exclusions, and indexes source documents, static files, components, layouts, runtime-template names, and hooks.

Components and layouts are indexed by name but not parsed eagerly (they are lazy). A broken unused component can therefore remain dormant; once used, it is very much your problem.

## 2. Pre-build hooks

The temporary and generated workspaces are cleared, one Lua runtime is created for the build, and pre-build hooks execute alphabetically. Hooks can prepare data or assets and call `documents.render` to register generated documents.

Registration does not write HTML. It adds a request to the same document plan used by source files.

## 3. Output planning

Source documents, static files, and generated requests claim their final paths. Generated requests are normalised and sorted. Traversal and collisions fail here, before Sklair removes the previous output.

## 4. Definition preparation

Source and generated bodies are parsed. Static inline Lua runs while each used definition is prepared. Component and layout definitions remain cached for reuse.

## 5. Instantiation and resolution

Bindings apply supplied props, dynamic Lua executes for the particular instance, component bodies are projected into slots, and recursive dependencies are resolved depth-first. Head contributions, runtime-template requirements, and used component folders propagate with the resolved graph.

Generated document bodies are compiled normally, then projected into an instantiated full-document layout.

## 6. Finalisation

Sklair registers native runtime templates when requested, segments and deduplicates the head, orders segments by browser significance, applies FOUC support when configured, and adds generator metadata. Development-server builds also receive their reload client.

## 7. Replacement and post-build work

The output directory is replaced, compiled documents and component assets are written, static inputs are copied, and only then do post-build hooks run. Post hooks can inspect and alter `built:` because the current output finally exists.

This order is deliberate. A pre-hook can decide what should exist; a post-hook can react to what was successfully built.
