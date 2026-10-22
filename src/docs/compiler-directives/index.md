{
  "description": "Control document head processing with ordering barriers and removal regions written as HTML comments."
}
---

# Compiler directives

Sklair recognises two paired HTML-comment directives in document heads:

- [Ordering barriers](./ordering-barriers.md) keep a group of head nodes together at a chosen sorting priority.
- [Removal regions](./removal-regions.md) exclude a group of head nodes from output.

Directive marker comments disappear during compilation. For how they interact with ordinary head elements, see [Head processing](../sklair/head-processing.md).
