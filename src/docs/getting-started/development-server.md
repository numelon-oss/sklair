{
  "description": "Build, serve, and automatically refresh a Sklair project while changing its source."
}
---

# Development server

Run this from the directory containing `sklair.json`:

```bash title="Terminal"
sklair serve
```

Sklair builds the project into a temporary directory, serves it locally, and watches the source, component, and layout directories. Connected pages refresh after a successful rebuild.

The development build contains a small reload client. A normal `#!bash sklair build` does not include it of course.

## Choose the address

Use `--port` to request a port and `--host` to change the listening address:

```bash title="Terminal"
sklair serve --host 0.0.0.0 --port 1234
```

| Flag                    | Default            | Purpose                                            |
| ----------------------- | ------------------ | -------------------------------------------------- |
| `#!bash --host`         | `#!text localhost` | Address on which the server listens.               |
| `#!bash --port`         | `#!text 0`         | Requested port. Zero chooses an available port.    |
| `#!bash --auto-refresh` | `#!text true`      | Refresh connected pages after successful rebuilds. |

Boolean values use Sklair's ordinary flag syntax:

```bash
sklair serve --auto-refresh=false
```

<!-- TODO: -open and -watch -->

!!! warning "Rebuild failures"

    If a rebuild fails, Sklair prints the compiler error and exits the development server. Fix the source and run `#!bash sklair serve` again.

<!-- TODO: keep the server alive and serve a Sklair error page from Go's embedded filesystem -->
