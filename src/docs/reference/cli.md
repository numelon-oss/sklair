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
| `clean`   | Remove configured output plus `.sklair/temp` and `.sklair/generated`. Persistent `.sklair/cache` data is retained. |
| `config`  | Attempt to open Sklair's global configuration file in the system editor.                                           |
| `docs`    | Open `https://sklair.numelon.com/docs` in the default browser.                                                     |
| `version` | Print the current Sklair version.                                                                                  |

`build` and `serve` find `sklair.json` in the current directory, then fall back to `src/sklair.json`. Command execution should therefore begin at the project root.

## Development server

```sh
sklair serve -host localhost -port 1234 -auto-refresh=true
```

`-port 0` selects an available port. `-auto-refresh` controls whether connected previews reload after a successful rebuild. The server currently rebuilds the complete project after any watched change.

The accepted `-open` and `-watch` flags are not wired into behaviour yet: the browser is not opened automatically, and watching remains enabled. They exist on the current CLI but should not be mistaken for completed features.

## Global options

| Option     | Effect                             |
| ---------- | ---------------------------------- |
| `-silent`  | Suppress everything except errors. |
| `-verbose` | Include informational output.      |
| `-debug`   | Include debug output.              |
| `-help`    | Show command help.                 |

`-silent` cannot be combined with `-verbose` or `-debug`. Global options must appear before the command.
