{
  "description": "Bundle a component entry document with its styles, modules, images, and other relative assets."
}
---

# Component folders and assets

A component folder keeps markup beside the resources it owns:

```text
components/
└── AudioTier/
    ├── index.html
    ├── audio-tier.css
    ├── audio-tier.mjs
    └── reference.svg
```

The directory name becomes the component name and `index.html` becomes its entry. A directory without `index.html` is ignored during component discovery.

## Declaring assets

Assets remain explicit HTML dependencies:

```html
<!doctype html>
<html>
<head>
  <link rel="stylesheet" href="./audio-tier.css">
  <script type="module" src="./audio-tier.mjs"></script>
</head>
<body>
  <span class="audio-tier">
    <img src="./reference.svg" alt="">
    <span sklair-text="label"></span>
  </span>
</body>
</html>
```

Since in most cases the source contains relative paths, and Sklair preserves the directory structure for component folders in the final site, assets should work normally. It doesn't matter though; read below.

## Built namespace

When the component is used, its directory is copied to:

```text
/_sklair/components/audiotier/
```

Relative references in the component's resolved head and body are rebased to that namespace. Root-relative, fragment-only, protocol-relative, and absolute URLs remain unchanged. Query strings and fragments on relative URLs are preserved.

Rewriting covers `href` on `<link>`, `src` and `srcset` on common media and embedded elements, and `data` on `<object>`. A relative resource path which escapes the component directory with `..` is rejected.

## Internal relative imports

The whole directory is copied with its internal structure intact. CSS `url()` values and JavaScript imports are not rewritten by the HTML compiler, but they continue to work when they are relative to the copied CSS or module file.

Only component folders reached by a compiled document are copied. Their assets propagate through recursive dependencies and are deduplicated by component name for the final build.
