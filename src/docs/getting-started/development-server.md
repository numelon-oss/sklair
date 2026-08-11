{
  "description": "Use Sklair's development server to rebuild and refresh a site while working on it."
}
---

# Development server

`sklair serve` builds the current project into a temporary output directory, serves it locally, watches project inputs, and rebuilds after changes.

```sh
sklair serve --port 8080
```

The development build receives a small reload client which is not present in ordinary `sklair build` output. After a successful rebuild, connected pages refresh automatically.

## Useful flags

```text
-host string
    Host to listen on. Default: localhost.

-port int
    Requested port. Zero asks Sklair to acquire an available port.

-auto-refresh bool
    Refresh connected previews after rebuilds. Default: true.
```

Sklair's flag syntax accepts forms such as `--auto-refresh=false`.

<!-- TODO: -open and -watch -->

## What is watched

The input, component, and layout directories are watched. Paths already beneath `input` are covered by its recursive watcher; configured directories outside it receive their own watcher, with each event currently triggering a complete build.

If a rebuild fails, the server command reports the compiler error and exits. Fix the source and start it again.
<!-- TODO: nice sklair error pages. for this at that point just use go's embedded filesystem to store all our runtime and ws stuff temporarily including the error page -->

A custom port can be specified as follows:

```sh
sklair serve --host 0.0.0.0 --port 1234
```
