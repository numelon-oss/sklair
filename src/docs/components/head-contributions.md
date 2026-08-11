{
  "description": "Let body-invoked components contribute deduplicated and ordered nodes to the final document head."
}
---

# Component head contributions

A component is invoked from document content even when its visible body is empty. Its `<head>` is still allowed to contribute metadata, styles, scripts, fonts, or other document-level dependencies.

```html
<!-- components/SiteHead.html -->
<!doctype html>
<html>
<head>
  <link rel="preconnect" href="https://fonts.example">
  <link rel="stylesheet" href="/styles/global.css">
</head>
<body></body>
</html>
```

```html
<body>
  <SiteHead></SiteHead>
  <main>...</main>
</body>
```

Components are not invoked inside `<head>`. Keeping invocation in the body avoids the ambiguous question of whether a component's body should become head content, disappear, or be teleported somewhere based on random nothingness. This has remained an "issue" for a long time now, but genuinely for the better.

## Contribution order

Recursive dependencies contribute before their owner. If `Player` uses `AudioTier`, the dependency's head is appended before `Player`'s head. Source traversal then continues with the next invocation.

The same resolved component instance contributes at most once per document. Instances with different prop signatures may produce different head nodes and are therefore considered separately.

## Final head processing

<!-- TODO: write separate page about head processing, including the processing order -->

After component resolution, Sklair segments the complete head. Ordinary duplicate nodes are removed by a structural weak hash and segments are stably ordered according to browser loading significance.

Ordering-barrier segments keep their internal order and are not deduplicated node by node. Use them only where an actual dependency sequence exists.

Head bindings and dynamic Lua use the component instance's props just like body markup. Runtime template heads are different: they are contributed statically and cannot contain runtime bindings.
