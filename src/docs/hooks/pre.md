{
  "description": "Prepare data, generated assets, and document requests before Sklair freezes and compiles its build plan."
}
---

# Pre-build hooks

Pre-build hooks run after discovery and workspace preparation, but before Sklair freezes its output plan. This is the phase for turning source data into planned compiler work.

A pre-hook can:

- read project inputs and persistent cache data;
- write temporary or generated files;
- parse Markdown or validate structured content;
- register any number of documents through `documents.render`.

It cannot inspect `built:` because the new output does not exist. Reading an old build during planning would also make clean builds and incremental-looking builds disagree, which is not a particularly charming property for a compiler.

Each `.lua` file in `hooks/pre` runs alphabetically in its own scope. A Lua error or non-zero `os.exit` stops the build. `os.exit(0)` ends only the current hook successfully and allows the next hook to continue.

Files written beneath `generated:` are copied to `/_sklair/` in the final output after compilation. Temporary files remain build-local and are cleared before the next build; cache files persist.
