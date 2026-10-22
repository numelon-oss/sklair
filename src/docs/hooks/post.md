{
  "description": "Inspect or modify completed output after Sklair has compiled documents and copied static and generated assets."
}
---

# Post-build hooks

Post-build hooks run only after Sklair has replaced the output directory, written compiled documents, copied static files and component assets, and installed generated assets.

This is the first phase with access to `built:`. A post-hook can inspect or write completed output - for example, producing a manifest which depends on final filenames or integrating with a tool that consumes the assembled tree.

Post hooks do not receive the `documents` module. The document plan is already frozen and compiled; adding another request here would mean rebuilding the build from inside its final phase.

Within a post hook, `generated:` refers to the generated asset subtree already copied beneath `built:/_sklair/`. Writes to either namespace affect the current output.

See [Filesystem](../lua/filesystem.md) for the complete namespace permission table.
