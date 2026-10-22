# Release packaging

Packaging for all the sklair releases.

- Manual runs accept a version to create a `release-bundle` artefact w/o creating a GitHub release
- Pushed `v<version>` tag also publishes that bundle, using `<tag>-RELEASE.md` for its release notes
- macOS and Windows cross compiled from linux
- APT signing and package manager metadata publication belongs to Numelon OSS Distribution repo

## Local packaging

Use Go from `go.mod`, `tar`, and `zip`. Linux packaging also needs nFPM:

```sh
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.48.0
bash .github/workflows/packaging/package.sh 0.4.0 darwin arm64
bash .github/workflows/packaging/package.sh 0.4.0 linux amd64
```

- Optional fourth arg selects an output directory instead of `dist`
- `SKLAIR_COMMIT`, `SKLAIR_BUILD_DATE` overrides for bin

## Artefacts

| Name                                 | Contents                                                           |
| ------------------------------------ | ------------------------------------------------------------------ |
| `sklair-darwin-{amd64,arm64}.tar.gz` | `sklair`, `LICENSE`, `README.md`                                   |
| `sklair-linux-{amd64,arm64}.tar.gz`  | `sklair`, `LICENSE`, `README.md`                                   |
| `sklair-windows-{amd64,arm64}.zip`   | `sklair.exe`, `LICENSE`, `README.md`                               |
| `sklair_<version>_{amd64,arm64}.deb` | `/usr/bin/sklair` and documentation under `/usr/share/doc/sklair/` |
| `sklair-v<version>.schema.json`      | Configuration schema with a version-specific identifier            |
| `sklair.schema.json`                 | The same schema under a stable filename                            |
| `SHA256SUMS`                         | SHA-256 checksums of all the above files                           |

- debian filenames contain their package version
  - a prerelease such
as `0.4.0-rc.1` becomes debian version `0.4.0~rc.1` so that it sorts before `0.4.0`
