# Sklair

Sklair is a HTML compiler and templating engine which makes building sites with HTML, CSS, and JavaScript both elegant and powerful again, through reducing repetition for large sites.

## The Sklair Philosophy

Using this compiler requires some context as to why it was made in teh first place.

Sklair is essentially a protest against modern frontend excess.

Sklair parses HTML - not JSX - and it is not a full framework and doesn't inject a bunch of JavaScript to create virtual DOM trees. It injects logic **only where needed** through (future) compiler directives and embedded Lua, uses plain `.html` files as component boundaries, has zero runtime dependencies (no hydration, no runtime VDOM, no `bundle.js`), and actually treats the web as a markup-first pplatform, not a javascript rendering target.

Therefore, Sklair is not just a tooling choice, but rather an ideological revolt against complex build steps, 800 kB "starter kits", component trees that only exist in memory, and, JS controlling the entire DOM lifecycle. Instead it says: "What if your website was literally just a folder of HTML, CSS and JS, but with just enough compiler help to stay dry, clean, and powerful?". It is a return to native HTML as the primary language, CSS *actually* doing layour instead of being second-class (although we at Numelon prefer to use Tailwind CSS, ironically), and JavaScript added for interactivity, not absolutely everything.

In summary, Sklair is basically old-school but modernised static site generation.

## Real use cases

In all truth, Sklair targets quite a niche audience. But lets theoretically say that you are a developer (or develoopers) who absolutely despise every single framework and refuse to use the likes of React, NextJS, et cetera. You like using regular HTML and actual CSS (or TailwindCSS) to make your sites not look like garbage from the 1990s. And on top of that, the only JavaScript in your site *is* actually functional and serves a purpose (i.e. to make a button do work, add animations, etc.) - to the point where it is genuinely a single page app.

In this case, you will have a single `.html` file that is ***absolutely huge*** and almost innavigable, and whilst you hate frameworks, you must admire the utility of being able to write `<CustomComponent />` in common frameworks and it saves you some time and provides you with consistency, the ability to update the same exact component across many pages where you use it.

Another use case is that you simply maintain an absolutely large corporate website where you need site consistency and really cannot afford to spend time copy and pasting a new link into the menu section of a site where that same menu bar is repeated across 1000 files.

Sklair is for you.

## How does it work?

1. Sklair performs document discovery by recursively finding HTML and static files in your project.
2. The `components` directory of your specified source is then scanned, with each file (hereunto referred to as a "component") being parsed - this is component discovery.
3. On-demand caching is performed: if a component is static (i.e. there are no Lua directives in it), then it is parsed and cached immediately. Otherwise if dynamic, its Lua blocks are kept for runtime processing per individual HTML file.
4. For each file, non-standard HTML tags are replaced with matching components. Output is written to a mirrored `build` directory structure.
5. Non-HTML assets are copied as-is into `build`.

## Performance considerations

Somewhat decent performance considerations have been put into Sklair.
I initially though of doing component discovery and file discovery sequentially, but then I realised that we would be scanning a bunch of files multiple times and parsing their contents even if we didnt need some components.
Therefore, there is lazy loading and whatnot. Below describes it in better detail:

- Lazy component caching means components are parsed only when they are actually used in your source HTML files
- Static versus dynamic resolution is performed, where Lua components are handled separately for performance
- Once parsed, component trees are reused so that there is no repeated parsing
- There is no lookup cost for routing or output structuring

## Example

> [!NOTE]  
> Component names are case-insensitive. You may choose to name a component `SomeHeader` for clarity, but Sklair treats both `someheader` and `SomeHeader` as the same.
> 
> Likewise, component files are case-insensitive.
> 
> You may choose to write `SomeHeader` in your HTML, and have your component saved as `someheader.html`, and it will still work.

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

### Compiled output

```html
<body>
  <header>
    <h1>Welcome to my site</h1>
  </header>
  
  <p>...</p>
</body>
```