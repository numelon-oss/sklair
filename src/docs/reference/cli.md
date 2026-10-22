{
  "description": "Look up Sklair's build, serve, init, and clean commands and global logging options."
}
---

# Command-line interface

Sklair uses a command followed by command-specific arguments:

```sh
sklair [global options] <command> [command options]
```

## Commands

| Command   | Behaviour                                                                                                          |
| --------- | ------------------------------------------------------------------------------------------------------------------ |
| `build`   | Compile the configured project into its output directory.                                                          |
| `serve`   | Build into a temporary directory, serve it, watch project inputs, and rebuild on change.                           |
| `init`    | Interactively create a project configuration in the current directory.                                             |
| `clean`   | Remove configured output and the entire `.sklair` workspace, including hook-created caches.                        |
| `docs`    | Open `https://sklair.numelon.com/docs` in the default browser.                                                     |
| `version` | Print the current Sklair version.                                                                                  |

`build` and `serve` find `sklair.json` in the current directory, then fall back to `src/sklair.json`. Command execution should therefore begin at the project root.

## Help and arguments

```bash
sklair --help
sklair help serve
sklair serve --help
```

`-h` is also accepted. Command help lists its options and examples without running the command or loading a project configuration.

Commands currently accept no positional arguments. Unknown flags and unexpected arguments are usage errors. Options must appear before any positional arguments, following Sklair's flag syntax.

Exit codes are `0` for success or requested help, `1` for an operation which failed, and `2` for incorrect usage. Running Sklair without a command prints help and exits with `2`.

## Development server

```sh
sklair serve -host localhost -port 1234 -auto-refresh=true
```

`-port 0` selects an available port. `-auto-refresh` controls whether connected previews reload after a successful rebuild. The server currently rebuilds the complete project after any watched change.

`--open` opens the preview after the initial successful build. `--watch=false` serves that build without watching for changes. See [Development server](../getting-started/development-server.md) for the full option table.

## Global options

| Option     | Effect                             |
| ---------- | ---------------------------------- |
| `-silent`  | Suppress everything except errors. |
| `-verbose` | Include informational output.      |
| `-debug`   | Include debug output.              |
| `--help`, `-h` | Show command help.             |

`-silent` cannot be combined with `-verbose` or `-debug`. Global options must appear before the command.
