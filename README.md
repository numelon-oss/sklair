<table>
<tr>
<td width="140" align="center">

<img src="https://sklair.numelon.com/branding/icon-colour.svg" width="120"/>

</td>
<td>

# What is Sklair?  

**Sklair is a HTML compiler.**

It takes real HTML files, reusable components, a few compiler directives, and pre/post-build hooks written in Lua, and produces better HTML than a human could realistically maintain by hand.

<p>
<a href="https://sklair.numelon.com" style="text-decoration:none;">
  <img alt="Website" height="40" src="https://cdn.jsdelivr.net/npm/@intergrav/devins-badges@3/assets/cozy/documentation/website_vector.svg">
</a>
<!-- &nbsp;
<a href="https://sklair-docs.numelon.com">
  <img alt="Read the documentation" height="40" src="https://cdn.jsdelivr.net/npm/@intergrav/devins-badges@3/assets/cozy/documentation/readthedocs_vector.svg">
</a> -->
&nbsp;
<a href="https://github.com/sponsors/numelon-oss" style="text-decoration:none;">
  <img alt="GitHub Sponsors" height="40" src="https://cdn.jsdelivr.net/npm/@intergrav/devins-badges@3/assets/cozy/donate/ghsponsors-plural_vector.svg">
</a>
</p>

</td>
</tr>
</table>

## What Sklair gives you

- **HTML components**
- **Head deduplication with a [heuristic head ordering pass](#compiled-output-example-2)**
- Social metadata generation (OpenGraph, Twitter)
- **Automatic resource hinting (preconnect, dns-prefetch, etc.)**
- **Compiler directives for advanced control**
- A live development server (`sklair serve`)
- Utilities to **prevent FOUC** (Flash Of Unstyled Content)
- Zero runtime JavaScript overhead, because this isn't a framework

All while outputting plain, static HTML.

### Example 1 - Simple

> [!NOTE]  
> Component names are case-insensitive. You may choose to name a component `SomeHeader` for clarity, but Sklair treats both `someheader` and `SomeHeader` as the same.
>
> Likewise, component files are case-insensitive.
>
> You may choose to write `SomeHeader` in your HTML, and have your component saved as `someheader.html`, and it will still work.

#### Source (Example 1)

```html
<!-- src/index.html -->
<body>
  <SomeHeader></SomeHeader>

  <Content></Content>
</body>
```

```html
<!-- components/SomeHeader.html -->
<header>
  <h1>Welcome to my site</h1>
</header>
```

```html
<!-- components/Content.html -->
<p>...</p>
```

#### Compiled output (Example 1)

```html
<body>
  <header>
    <h1>Welcome to my site</h1>
  </header>

  <p>...</p>
</body>
```

### Example 2 - Components with head + body insertion

#### Source (Example 2)

```html
<!-- src/index.html -->
<!DOCTYPE html>
<html>

<head>
    <title>Hello world</title>
</head>

<body>
    <CommonHead></CommonHead>

    <Content></Content>
</body>

</html>
```

```html
<!-- components/CommonHead.html -->
<!DOCTYPE html>
<html>
    <head>
        <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">

        <script src="/assets/js/ThemeSync.js" defer></script>
        
        <!-- sklair:ordering-barrier treat-as=script -->
        <script src="https://cdn.tailwindcss.com"></script>
        <script src="/assets/styles/tailwind.config.js"></script>
        <link rel="stylesheet" href="/assets/styles/global.css">
        <link rel="stylesheet" href="/assets/styles/themes.css">
        <!-- sklair:ordering-barrier-end -->
        
        <link rel="preconnect" href="https://wcdn.numelon.com">
        <meta charset="UTF-8">
        <meta name="theme-color" content="#ff4e4e">
        <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
    </head>

    <body>
    </body>
</html>
```

```html
<!-- components/Content.html -->
<p>...</p>
```

#### Compiled output (Example 2)

As can be seen below, although the `<head>` of the `CommonHead` component was unordered above and semantically chaotic, the compiled output is deterministically reordered according to browser loading semantics.

**This is not a cosmetic change**, albeit looking nicer is a plus. Modern browsers stream-parse HTML as the bytes for it arrive, and many `<head>` elements (such as `meta charset`, preconnects, stylesheets and scripts) trigger side effects the moment they are encountered. Their relative position therefore directly affects request scheduling, render-blocking, and even URL resolution.

Sklair applies a heuristic head-ordering pass to ensure that these high-impact nodes are discovered _as early as possible_, improving page load behaviour without requiring authors to manually micromanage this order themselves.

Additionally, the use of the `sklair:ordering-barrier` compiler directive has allowed us to preserve the order of HTML elements where it may be important, thus preventing the break of some things during the re-ordering pass. For example, the most common use case we have encountered is ensuring that Tailwind configurations load after the tailwindcss script itself.

```html
<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8" />

    <meta name="theme-color" content="#ff4e4e" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />

    <link rel="preconnect" href="https://wcdn.numelon.com" />

    <title>Hello world</title>

    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&amp;display=swap"
        rel="stylesheet" />
    <script src="/assets/js/ThemeSync.js" defer=""></script>

    <script src="https://cdn.tailwindcss.com"></script>
    <script src="/assets/styles/tailwind.config.js"></script>
    <link rel="stylesheet" href="/assets/styles/global.css" />
    <link rel="stylesheet" href="/assets/styles/themes.css" />

    <meta name="generator" content="https://sklair.numelon.com" />
</head>

<body>
    <p>...</p>
</body>

</html>
```

## How does it work?

1. Pre-build Lua hooks run, if declared in `sklair.json`
    - These hooks have the ability to write to `.sklair/generated`, `.sklair/tmp` and `.sklair/cache`, which are also available to post-build hooks.
2. Sklair scans your project for HTML and static assets
3. It discovers all components in your components directory
4. Components are parsed lazily only when needed
5. Non-standard tags are replaced with components
6. `<head>` is analysed, deduplicated, and heuristically "optimised"
7. Processed (built) HTML files are written into a build directory, and original static files are copied verbatim
    - Files from `.sklair/generated` are copied to `_sklair/generated` inside the build directory
8. Post-build Lua hooks run, if declared in `sklair.json`
    - These hooks also have the ability to read and write files inside the build directory

## Performance

### The tool itself

- Components are parsed once and cached
- Static components are reused across files
- Dynamic components are evaluated only when needed

Apart from these, there are a few other optimisations which make Sklair have tiny compile times, even for big projects with many components and many usages.

### Compiled output in browsers

This is where the ideological part comes in. The web sites produced with Sklair _should_ be fast because of the way it forces you to structure your project and think about development without frameworks, routing, or hydration - you implement only what you really need.

## Who is this for?

Sklair was originally written as a proprietary tool for use within Numelon, so, we have a genuine use for this tool, but in general Sklair is for people who:

- want **reusable components** without rewriting the web in some nasty framework
- maintain large websites and are tired of copy-pasting headers, navigation bars, and footers across hundreds of files
- need **composability** for HTML
- want to build light SPAs using HTML + JS instead of virtual DOMs (and yes, it's possible to make SPAs with Sklair)

## The philosophy behind Sklair

The real problem is not frameworks, but rather the slow destruction of HTML as a document format.

Many modern tools (JSX, Jinja, Liquid, and countless templating DSLs) stop treating HTML as HTML itself. They turn it into a syntax tree, a string generator, or JavaScript mixed with markup in an unusually messy way.

Markup, logic, and control flow become entangled inside curly braces, percent blocks, and pseudo-languages that vaguely resemble HTML. Even when these tools output HTML in the end, the authoring experience has already broken the contract - the source document no longer _looks_ like a HTML document.

Sklair deliberately refuses to do this. Instead:

- Sklair keeps HTML looking like HTML. The semantics of HTML are preserved, and, in fact, Sklair will not parse semantically incorrect source documents.
- Components are written as real tags (`<SomeComponent></SomeComponent>`), not curly-brace expressions, JSX-style tags, or something else
- Compile-time logic lives in dedicated `<lua></lua>` blocks, with some additional guidance from compiler directives in comments (`<!-- sklair:... -->`) rather than being smeared across every line of markup.

This therefore creates a clean separation between HTML as a markup language, compile-time logic, and runtime JavaScript staying runtime JavaScript. Nothing is pretending to be something it is not.

This also naturally encourages a healthier structure for sites. Because Sklair is a compiler rather than a framework, you are pushed toward having real pages, real documents, and real navigation - or, if you choose, hidden components that you can animate and reveal yourself - instead of a single giant page pretending to be an entire website through JavaScript and also behind a fake address bar.

In other words, Sklair does not try to replace the browser's model of the web. It embraces it, and simply gives you better tools to work with it.

### You're so against React? Frameworks? What’s wrong with you? (Note on SPAs)

This is not about forbidding frameworks or banning JavaScript, it is about refusing to lie to the browser about what a document is.

Framework-style SPAs often simulate navigation, URLs, and pages on top of a single HTML file, even though the browser itself is already built around real documents and real navigation. Ironically, these systems still need multiple HTML entry points just to route everything back into that one fake page - defeating the very illusion they try to create with SPAs.

With sklair, **you can absolutely build SPAs**. You just do it more honestly by using:

- JavaScript modules
- Regular DOM APIs, no VDOM
- Animations
- State stored in real variables, not `useState()` etc

You can have overlays, modals, panels, dashboards, and transitions, exactly like a framework - except they already exist in the HTML and are revealed, animated, or updated by your own code, instead of being created and destroyed inside a virtual DOM.

<!-- TODO: add that other new project to this list as an example? should it be public? -->
In fact, that is how [Numelon Passport](https://numelon.fandom.com/wiki/Numelon_Passport) was built - with no hydration, JSX, or virtual DOM. Simply documents, components, and JavaScript doing exactly what it needs to do.

Sklair does not stop you from building rich applications, it stops you from pretending that your application _isn't_ made of documents.

## License

Sklair is licensed under AGPL-3.0.

**User‑provided content remains the property of the user**. The user owns output generated by Sklair based on user‑provided content, provided that such output **does not contain copyrighted material from Sklair itself**.

The copyright holder hereby releases all commits made prior to the introduction of the AGPL‑3.0 licence under the same AGPL‑3.0 licence.
