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

!!! warning "Rebuild failures"

    Do note that binding to `0.0.0.0` in this particular example listens on all IPv4 interfaces, allowing other devices on your local network to connect using your computer's LAN address. That is useful for testing on another device like a phone, but you should still keep this in mind.

    The default host however, as stated below, is `localhost`.

| Flag                    | Default            | Purpose                                                                    |
| ----------------------- | ------------------ | -------------------------------------------------------------------------- |
| `#!bash --host`         | `#!text localhost` | Address on which the server listens.                                       |
| `#!bash --port`         | `#!text 0`         | Requested port. Zero chooses an available port.                            |
| `#!bash --auto-refresh` | `#!text true`      | Refresh connected pages after successful rebuilds.                         |
| `#!bash --open`         | `#!text false`     | Open the preview in your default browser after the first successful build. |
| `#!bash --watch`        | `#!text true`      | Watch project files and rebuild when they change.                          |

Boolean values use Sklair's ordinary flag syntax:

```bash title="Serving without automatic refresh"
sklair serve --auto-refresh=false
```

Use `--watch=false` to serve a single build without watching files. `--auto-refresh=false` keeps rebuilding while leaving connected pages alone. Press Ctrl+C to stop the server and remove its temporary output.

!!! warning "Rebuild failures"

    If a rebuild fails, Sklair prints the compiler error and exits the development server. Fix the source and run `#!bash sklair serve` again.

<!-- TODO: keep the server alive and serve a Sklair error page from Go's embedded filesystem -->
