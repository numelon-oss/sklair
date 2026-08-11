{
  "description": "See why Sklair treats ordinary HTML as its final product rather than an intermediate runtime representation."
}
---

# Bare HTML output

Bare output means the browser receives the final document rather than a serialised promise that JavaScript will construct one later.

Ordinary component tags disappear. Compile-time bindings disappear. Inline Lua scripts disappear. Slots disappear. The output contains the HTML nodes those authoring facilities produced, plus the CSS and JavaScript the site deliberately included.

## What Sklair does not ship

An ordinary build has no component registry, hydration pass, virtual DOM, expression evaluator, or client-side lifecycle. There is no need for the browser to rediscover work the compiler completed while it had the whole project available.

This does not forbid runtime JavaScript. A music player still needs playback state; a theme control still needs to respond to a click. Sklair merely declines to make every paragraph participate in that runtime because one button somewhere is interactive.

## The deliberate exception

Configured [runtime templates](/docs/components/runtime-templates/) produce a small ES module and native `<template>` registry. They exist for markup which genuinely must be instantiated after page load - for example, a search result created after a query or an audio-quality badge created from newly loaded track data.

Even then, the runtime clones `DocumentFragment` objects and applies a finite set of bindings once. It is **NOT reactive** and it does not take ownership of the surrounding document.

## Why this matters

Bare output is inspectable, cacheable, deployable to any static host, and compatible with the browser's own navigation and accessibility model. More importantly, it keeps the abstraction bill attached to authoring rather than charging every visitor indefinitely.

<!-- TODO: examples here being numelon passport, rubis v3 frontend, this entire documentation site, but again this is a TODO for LATER. perhaps a dedicated exxamples page later -->

Sklair can build substantial SPAs. It simply expects their JavaScript to use the platform honestly: modules, DOM APIs, events, variables, and whatever application state the application actually has.
