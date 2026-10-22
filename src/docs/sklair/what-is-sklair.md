{
  "description": "Build websites and browser applications with HTML components, Lua, document generation, and optional runtime templates."
}
---

# What is Sklair?

Sklair is a HTML compiler. More specifically, it is an ahead-of-time, source-to-source compiler for a HTML-based authoring language, with embedded Lua evaluation and a document generation pipeline.
You write HTML documents and reusable components; Sklair combines them during a build and writes a static website.

Components can have props, conditional content, bodies, styles, scripts, and head contributions. Component tags, bindings, slots, and compile-time Lua disappear from the result, leaving the HTML they produced.

```html title="Source document"
<SiteHeader></SiteHeader>
<ReleaseCard title="Version 2.4"></ReleaseCard>
<SiteFooter></SiteFooter>
```

The output is ordinary HTML, CSS, JavaScript, and other assets. You can deploy it to a static host. Sklair does not require hydration, a virtual DOM, or a component lifecycle in the browser.

## Lua and document generation

[Lua](../lua/index.md) can run inside documents, components, and layouts during compilation. It can read data, calculate values, and emit markup. [Build hooks](../hooks/index.md) run Lua before or after the build to prepare inputs or work with completed output.

[Document generation](../document-generation/index.md) lets a hook register multiple documents using full-document layouts. A directory of Markdown articles or a collection of JSON records can therefore become an entire website. This documentation is generated with those facilities.

## Browser JavaScript

Include JavaScript whenever your site needs browser behaviour, such as search, audio playback, or a theme control. Sklair can build substantial SPAs as well as websites with normal page navigation.

For markup created after page load, [runtime templates](../components/runtime-templates.md) provide an optional small ES module. It clones native templates and fills their bindings once. It is not reactive and does not manage the surrounding document.
