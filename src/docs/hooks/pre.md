{
  "description": "Prepare data, write generated assets, and register documents before compilation."
}
---

# Pre-build hooks

Pre-build hooks run after discovery and workspace preparation, but before Sklair freezes its output plan. This is the phase for turning source data into planned compiler work.

A pre-hook can:

- read project inputs and persistent cache data;
- write temporary or generated files;
- parse Markdown or validate structured content;
- register any number of documents through `documents.render`.

It cannot inspect `built:` because the new output does not exist. See [Filesystem](../lua/filesystem.md) for the complete namespace permission table.

Files written beneath `generated:` are copied to `/_sklair/` in the final output after compilation. Temporary files remain build-local and are cleared before the next build; cache files persist.
