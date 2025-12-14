# Sklair Known Bugs

These are some insane bugs that can't be figured out yet with Sklair (unless some big changes are made)

## Component usage broken in source `<head>`

When placing a component (e.g. `<CustomAnalytics>`) inside the `<head>` tag of a source document, unexpected behaviour occurs due to the behjaviour of Go's `x/net/html` parser:

1. If the component is the last child inside `<head>`, it silently disappears from the output entirely
2. If the component has other sibling tags after it, the entire component and all subsequent tags are moved into the `<body>` of the final output, despite being authored inside `<head>`

This happens at **parse time** (Go's html parser - nothing to do with us!), not during replacement, meaning the tree is already malformed by the time we process it.

### Root cause

Go's `html.Parse` uses the HTML5 parsing rules, which follow browser logic:

- HTML components like `<script>`, `<style>`, `<title>`, etc are valid in `<head>`, but "unknown" tags (like `<h1>`, `<p>` - and, in this case, custom components) are treated as unexpected
- The parser assumes they are part of the body, and repositions them accordingly - before Sklair even gets to walk the tree

Therefore, if you write the following in a source file:

```html
<head>
    <CustomAnalytics></CustomAnalytics>
    <title>Hello bugs!</title>
</head>
```

it is parsed as such:

```html
<head></head>
<body>
    <CustomAnalytics></CustomAnalytics>
    <title>Hello bugs!</title>
</body>
```

Then, when the component is replaced with its content, it is insertedf into the body, not the head - and worse, tags that were after the component (e.g. `<title>`) also get moved into the body, completely violating the intended HTML structure.

### Workaround (its bad)

The simplest workaround for now is to simply just not use any components inside head tags. Yeah... not ideal, however, the "fix" for this issue is disproportionately costly compared to the urgency of having components inside head tags. We can simply just wait until a custom parser is written, using `x/net/html` as the tokenisation base.

### Custom parser info

A custom parser/tree builder override will probably be written to prevent `html.Parse` from reassigning unknown tags like `<Componennt>` into the wrong part of the DOM.
The most likely option right now is to quite literally just take the source code of `html.Parse`, modify it a bit to NOT put unknown tags from head into body (AT LEAST for components) and just use that.
