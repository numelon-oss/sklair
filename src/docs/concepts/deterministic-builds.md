{
  "description": "Learn which parts of a Sklair build are ordered, deduplicated, cached, and rejected when ownership is ambiguous."
}
---

# Deterministic builds

Determinism is less glamorous than a component demo and considerably more valuable once a project has hundreds of documents.

Sklair gives every output one owner, sorts work at the boundaries where order matters, deduplicates only where the rule is explicit, and rejects ambiguous graphs.

## Ordered inputs

Hooks run alphabetically within `pre` and `post`. Generated document requests are sorted by output path before compilation. Source and generated documents enter one output-sorted plan. Documentation generators and similar hooks can therefore compute a complete site before registration without inheriting directory iteration order by accident.

Recursive components use depth-first resolution in source order. This matches where an invocation appears in the document and makes dependency cycles report the actual chain which led back to the component.

## Output ownership

Source documents, copied static files, and generated documents may not claim the same output path. Generated outputs must be relative files beneath the build root and cannot use traversal, absolute paths, URI schemes, or trailing-directory names.

The plan freezes before compilation. A late registration cannot silently overwrite something which was already written.

## Head deduplication and ordering

Ordinary head elements are weakly hashed and duplicate nodes are removed. Ordering-barrier blocks remain intact and are not deduplicated as individual nodes. The remaining segments are stably ordered by browser significance: charset and HTTP metadata first, then viewport metadata and resource hints, title, styles, scripts, user-marked analytics, and unconstrained content.

The compiler orders resource hints you supplied; automatic hint invention is not implemented.
