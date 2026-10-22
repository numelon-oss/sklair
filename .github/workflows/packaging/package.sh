#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 3 || $# -gt 4 ]]; then
    echo "Usage: bash .github/workflows/packaging/package.sh <version> <os> <arch> [output-directory]" >&2
    exit 2
fi

version=${1#v}
target_os=$2
target_arch=$3
output_dir=${4:-dist}

if [[ ! $version =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$ ]]; then
    echo "Invalid release version: $1" >&2
    exit 2
fi
case "$target_os/$target_arch" in
    darwin/amd64|darwin/arm64|linux/amd64|linux/arm64|windows/amd64|windows/arm64) ;;
    *) echo "Unsupported release target: $target_os/$target_arch" >&2; exit 2 ;;
esac

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
mkdir -p "$output_dir"
output_dir=$(cd "$output_dir" && pwd)
nfpm_binary=${NFPM:-nfpm}
if [[ $target_os == linux ]]; then
    command -v "$nfpm_binary" >/dev/null || {
        echo "Linux packaging requires nFPM; install the version specified in the release workflow." >&2
        exit 1
    }
fi

package_tmp_dir=${TMPDIR:-/tmp}
package_work_dir=$(mktemp -d "${package_tmp_dir%/}/sklair-package.XXXXXX")
trap 'rm -r -- "$package_work_dir"' EXIT
binary_name=sklair
if [[ $target_os == windows ]]; then
    binary_name=sklair.exe
fi

commit=${SKLAIR_COMMIT:-$(git rev-parse HEAD)}
build_date=${SKLAIR_BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}
CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build \
    -mod=readonly -trimpath \
    -ldflags "-s -w -X sklair/constants.Version=v$version -X sklair/constants.Commit=$commit -X sklair/constants.BuildDate=$build_date" \
    -o "$package_work_dir/$binary_name" .

chmod 0755 "$package_work_dir/$binary_name"
install -m 0644 LICENSE "$package_work_dir/LICENSE"
install -m 0644 README.md "$package_work_dir/README.md"

# release URLs carry the version
# stable archive names also support GitHub's latest/download URLs.
archive_base="sklair-$target_os-$target_arch"
if [[ $target_os == windows ]]; then
    (cd "$package_work_dir" && zip -q -X "$archive_base.zip" "$binary_name" LICENSE README.md)
    mv "$package_work_dir/$archive_base.zip" "$output_dir/$archive_base.zip"
else
    COPYFILE_DISABLE=1 tar --no-xattrs -czf "$output_dir/$archive_base.tar.gz" -C "$package_work_dir" "$binary_name" LICENSE README.md
fi

if [[ $target_os == linux ]]; then
    # debian sorts '~rc.1' before a final release, '-' is a package-revision separator.
    deb_version=${version/-/\~}
    SKLAIR_PACKAGE_ARCH="$target_arch" \
    SKLAIR_PACKAGE_BINARY="$package_work_dir/$binary_name" \
    SKLAIR_DEB_VERSION="$deb_version" \
        "$nfpm_binary" package --packager deb --config .github/workflows/packaging/nfpm.yaml \
        --target "$output_dir/sklair_${deb_version}_${target_arch}.deb"
fi
