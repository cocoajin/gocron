#!/usr/bin/env bash
# Build a release archive without starting the application or creating user data.
set -euo pipefail
cd "$(dirname "$0")/.."
VERSION="$(tr -d '\r\n' < VERSION)"
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo 'VERSION must contain a three-part version.' >&2; exit 1
fi
if [[ -n "${RELEASE_TAG:-}" && "$RELEASE_TAG" != "$VERSION" ]]; then
  echo "Tag $RELEASE_TAG does not match VERSION $VERSION" >&2; exit 1
fi
export GOOS="${GOOS:-$(go env GOHOSTOS)}"
export GOARCH="${GOARCH:-$(go env GOHOSTARCH)}"
export CGO_ENABLED=1
case "$GOOS/$GOARCH" in
  linux/amd64|linux/arm64|windows/amd64|darwin/amd64|darwin/arm64) ;;
  *) echo "Unsupported target: $GOOS/$GOARCH" >&2; exit 1 ;;
esac
NAME="gocron-$VERSION-$GOOS-$GOARCH"
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT
mkdir -p "$STAGE/$NAME" dist
SUFFIX=''
if [[ "$GOOS" == windows ]]; then SUFFIX='.exe'; fi
COMMIT="$(git rev-parse HEAD)"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
LDFLAGS="-s -w -X main.AppVersion=$VERSION -X main.GitCommit=$COMMIT -X main.BuildDate=$BUILD_DATE"
# Bundle the C runtime on Linux/Windows; SQLite itself is compiled into the executable.
if [[ "$GOOS" == linux || "$GOOS" == windows ]]; then
  LDFLAGS="$LDFLAGS -linkmode external -extldflags -static"
fi
go build -tags netgo,osusergo -trimpath -ldflags "$LDFLAGS" -o "$STAGE/$NAME/gocron$SUFFIX" ./cmd/gocron
go build -tags netgo,osusergo -trimpath -ldflags "$LDFLAGS" -o "$STAGE/$NAME/gocron-node$SUFFIX" ./cmd/node
cp README.md LICENSE VERSION "$STAGE/$NAME/"
python3 - "$STAGE" "$NAME" "$GOOS" <<'PY'
import pathlib, sys, tarfile, zipfile
stage, name, target = sys.argv[1:]
root = pathlib.Path(stage) / name
if target == 'windows':
    output = pathlib.Path('dist') / (name + '.zip')
    with zipfile.ZipFile(output, 'w', zipfile.ZIP_DEFLATED) as archive:
        for item in sorted(root.iterdir()):
            archive.write(item, arcname=name + '/' + item.name)
else:
    output = pathlib.Path('dist') / (name + '.tar.gz')
    with tarfile.open(output, 'w:gz') as archive:
        archive.add(root, arcname=name)
print(output)
PY
