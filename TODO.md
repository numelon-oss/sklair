# Todo list

## todo list transferred from `numelon-proprietary/website` repo

- create numelon web packing tool for creating websites
  - this web packing tool should have hot reload
  - allow for components
    - e.g. `/components/someComponent.html` and `/components/someComponent.css`
    - numelon web packing tool will search for components dir
      - if html found then it will replace the tag inside the html with the actual content of the html file
      - if css found for component in component dir then it will be automatically copied to the build directory into css dir and also the link href stylesheet injected into the head of the html file where the component was used
      - if a component is not found, then just assume its a component registered from within js and leave it as-is, but issue a warning in the logs.
      - if js file is found then assume regular html component registered via js and add the js to head tag. ofc if css with same name is found then import that too
      - therefore, the order from most important to least is like this:
        1. componentName.js -> import js (& optional CSS, if HTML is present then ignored) in head
        2. componentName.html -> replace all occurrences of `<componentName/>` or `<componentName>` or `<componentName></componentName>` with contents of componentName.html. (& optional css in head)
            - note that if it is `<componentName>aaa</componentName>` then `aaa` would be passed to `$$COMPONENT_BODY` var inside componentName.html.
  - reusable component for menubar and footer
  - allow components to be used inside each-other (in src) but hard fail on circular component usage bc infinite loop

- support for replacing $$COMPONENT_BODY inside JavaScript too, since it's only supported in HTML right now with `<!-- $$COMPONENT_BODY -->`

- create a JSON schema for `sklair.json` files:
- <https://json-schema.org/understanding-json-schema/reference/index.html>
- create separate timers for actually processing the files - ie file discovery, then compiling. then separate timer for copying static files since that heavily inflates the build time.

## todo december 2025

- prepare for distribution to:
  - homebrew
  - winget
  - apt (self-hosted repo, because debian sponsorship is too slow)
  - regular GitHub releases -> links on website (although discourage using GitHub releases, make users use homebrew, winget, apt)
  - installation instructions on website very nice
- make sklair actually more of a cli tool
  - think in terms of subcommands:
    - `sklair init` -> creates a sklair.json file in the current directory and answer a questionnaire
    - `sklair build` -> builds the website based on sklair.json file or default values (if no sklair.json then warn about defaults available on docs)
    - `sklair serve` -> starts a local dev server, watching for changes and auto rebuild also ensure that its not actually built EVERY time theres a change - also make a preview page available at like /_.sklairpreview which allows you to preview what components look like independently
    - `sklair clean` -> removes all build artifacts (build dir, static dir)
    - `sklair update` -> updates sklair to the latest version (ALSO: ensure that on every run of sklair, it notifies the user of a new version unless auto update check is disabled in sklair config)
  - then only finally print a new empty line and then print build time stats etc (summary)
- search for "TODO" in the entire project and attempt to fix all of those
- ensure that in main.go the default sklair.json fallback is NOT src/sklair.json but rather just sklair.json. or just test both?
- long term: allow sklair to integrate third party stuff like tailwind compilation: sklair scans html, sees which classes are used, compiles css. likewise also scans css for tailwind class usage and adds them to css just in case, so that its also programmable.
- recursively parse components (whilst avoiding circular dependencies/components)
- allow components to be entire folders with index.html inside, and other files.
  - usage of a file from the component folder, e.g. the component index.html references the local style.css inside the component folder will be detected byy sklair and then will be rewritten once compiled so that all paths are not broken
  - at that rate, after compiling create a _sklair directory in the build folder where all component files live in and in the final html, component dependencies are referenced from there

## more todo (a bit long-term?)

sklair shouldn't just blindly replace components but also produce highly optimised output. yeah, sounds crazy for "just some HTML" but thats the point.

build optimisations:

- automatic resource discovery (eg external scripts, fonts, images) across documents and components (detect in final output)
- preconnect and dns-prefetch insertion (automatically after scanning source documents and components) - based on discovered external domains (eg fonts.googleapis.com), automatically insert optimised `<link rel="preconnect">` and `<link rel="dns-prefetch">` tags near the TOP of head
- heuristic based ordering of head tags for ideal performance
    1. preconnect/dns prefetch should be first
    2. stylesheets and scripts logically grouped
    3. meta and charset tags next
    4. analytics inserted last (will be very hard to actually define what analytics is - so simply detect from component name such as "analytics" or "tracking" or "google" etc)

- consider providing a final feedback summary at the end of compilation:
  - basically use all of the knowledge in web development thus far and try to provide it through sklair lol
  - "! consider self-hosting these common external dependencies to improve performance and reduce dns lookups" - sklair recommendation upon detecting common script tags or stylesheets etc (eg fontawesome from cloudflare cdnjs, fonts from google)

# documentation pages
- how to use sklair in github workflows (how to deploy to github pages)
- how to make a sklair website