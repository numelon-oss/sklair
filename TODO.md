# Todo list

## todo december 2025

- search for "TODO" in the entire project and attempt to fix all of those
- long term: allow sklair to integrate third party stuff like tailwind compilation: sklair scans html, sees which classes are used, compiles css. likewise also scans css for tailwind class usage and adds them to css just in case, so that its also programmable.

- in the future, all errors and warnings will have a link to the sklair documentation for more information

- favicon component similar to our builtin opengraph component
- actual sklair error pages instead of exiting the entire program on failure
- compile 404.html/.shtml and use that for custom 404 pages
  - when compiling 404.html/.shtml files, warn about relative links breaking (eg when a user visits /a/b/c/notfound instead of /notfound) - so always anchor to root OR give full paths to assets

## more todo (a bit long-term?)

sklair shouldn't just blindly replace components but also produce highly optimised output. yeah, sounds crazy for "just some HTML" but thats the point.

build optimisations:

- automatic resource discovery (eg external scripts, fonts, images) across documents and components (detect in final output)
- preconnect and dns-prefetch insertion (automatically after scanning source documents and components) - based on discovered external domains (eg fonts.googleapis.com), automatically insert optimised `<link rel="preconnect">` and `<link rel="dns-prefetch">` tags near the TOP of head
  - preconnect and dns prefetch must be inserted in the order that their respective domains are in the actual document. always head preconnect and dns prefetch first. but then after all that, if there is an image first at top of body from somecdn.com, then somecdn.com should be first preconnect and dns prefetch

- consider providing a final feedback summary at the end of compilation:
  - basically use all of the knowledge in web development thus far and try to provide it through sklair lol
  - "! consider self-hosting these common external dependencies to improve performance and reduce dns lookups" - sklair recommendation upon detecting common script tags or stylesheets etc (eg fontawesome from cloudflare cdnjs, fonts from google)

## documentation pages
- how to use sklair in github workflows (how to deploy to github pages)