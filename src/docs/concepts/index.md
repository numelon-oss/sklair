{
  "description": "Understand the ideas which make Sklair a compiler rather than a browser framework."
}
---

# Concepts

Sklair becomes much less surprising once it is treated as a compiler.

The source tree contains definitions and documents. A build discovers them, runs build-wide preparation, creates concrete instances, resolves their dependencies, finalises complete HTML documents, and replaces the output tree. Nothing continues managing those documents in the browser unless you explicitly requested a runtime template.

## The four ideas worth learning

[The compilation model](/docs/concepts/compilation-model/) explains the complete build in order. It matters when hooks, generated documents, static Lua, and post-build work begin interacting.

[Definitions and instances](/docs/concepts/definitions-and-instances/) explains why one component can be parsed once yet produce many prop-specific results, and why static and dynamic Lua have different lifetimes.

[Bare HTML output](/docs/concepts/bare-html-output/) explains the project's most opinionated constraint: authoring abstractions should disappear unless the browser genuinely needs them.

[Deterministic builds](/docs/concepts/deterministic-builds/) explains ordering, ownership, deduplication, and why ambiguous output fails instead of choosing a winner based on filesystem weather.
