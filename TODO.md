# Todo list

## todo list transferred from `numelon-proprietary/website` repo

- support for replacing $$COMPONENT_BODY inside JavaScript too, now just HTML with `<!-- $$COMPONENT_BODY -->`

## todo december 2025

- create a JSON schema for `sklair.json` files:
- <https://json-schema.org/understanding-json-schema/reference/index.html>

- prepare for distribution to:
  - homebrew
  - winget
  - apt (self-hosted repo, because debian sponsorship is too slow)
  - regular GitHub releases -> links on website (although discourage using GitHub releases, make users use homebrew, winget, apt)
  - installation instructions on website very nice
- make sklair actually more of a cli tool
  - think in terms of subcommands:
    - `sklair update` -> updates sklair to the latest version (ALSO: ensure that on every run of sklair, it notifies the user of a new version unless auto update check is disabled in sklair config)
  - then only finally print a new empty line and then print build time stats etc (summary)
- search for "TODO" in the entire project and attempt to fix all of those
- long term: allow sklair to integrate third party stuff like tailwind compilation: sklair scans html, sees which classes are used, compiles css. likewise also scans css for tailwind class usage and adds them to css just in case, so that its also programmable.

- extend CommandRegistry to allow per-subcommand help, also maybe fix per command flag parsing when required

- in the future, all errors and warnings will have a link to the sklair documentation for more information

- when compiling 404.html/.shtml files, warn about relative links breaking (eg when a user visits /a/b/c/notfound instead of /notfound) - so always anchor to root OR give full paths to assets

- create an icon component, similar to the opengraph one

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

- how does sklair work? (maintainability doc)
- how to use sklair in github workflows (how to deploy to github pages)
- how to make a sklair website

## for much later

<!-- - at some very late point, go through the entire project to see where we ARENT using pointers etc (avoid copying!!) and fix that -->
