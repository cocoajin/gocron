#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")/.."
export CGO_ENABLED=1
ARCH="$(go env GOARCH)"
if [ "$(go env GOOS)" != darwin ]; then echo 'Run this script on macOS.' >&2; exit 1; fi
if [ "${SKIP_WEB_BUILD:-0}" != 1 ]; then
  (cd web/vue && npm ci --ignore-scripts --no-audit --no-fund && npm run build)
  cp -R web/vue/dist/. web/public/
  go run github.com/rakyll/statik -src=web/vue/dist -dest=internal -f
  gofmt -w internal/statik/statik.go
fi
DEST="dist/gocron-sqlite-darwin-$ARCH"
APP="$DEST/GoCron.app"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources" bin
go build -trimpath -o bin/gocron ./cmd/gocron
go build -trimpath -ldflags '-X main.AppVersion=2.0' -o bin/gocron-node ./cmd/node
cp bin/gocron bin/gocron-node scripts/start.command "$APP/Contents/Resources/"
cat > "$APP/Contents/MacOS/GoCron" <<'LAUNCH'
#!/bin/bash
set -e
RESOURCES="$(cd "$(dirname "$0")/../Resources" && pwd)"
exec /usr/bin/open -a Terminal "$RESOURCES/start.command"
LAUNCH
cat > "$APP/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleExecutable</key><string>GoCron</string>
<key>CFBundleIdentifier</key><string>local.gocron.sqlite</string>
<key>CFBundleName</key><string>GoCron</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>2.0.0</string>
<key>CFBundleVersion</key><string>1</string>
<key>LSMinimumSystemVersion</key><string>12.0</string>
</dict></plist>
PLIST
chmod +x "$APP/Contents/MacOS/GoCron" "$APP/Contents/Resources/start.command"
cp LICENSE "$DEST/LICENSE"
cp README-SQLite.md "$DEST/使用说明.md"
codesign --force --sign - "$APP/Contents/Resources/gocron"
codesign --force --sign - "$APP/Contents/Resources/gocron-node"
codesign --force --sign - "$APP"
codesign --verify --deep --strict "$APP"
(cd dist && /usr/bin/ditto -c -k --sequesterRsrc --keepParent "gocron-sqlite-darwin-$ARCH" "gocron-sqlite-darwin-$ARCH.zip")
shasum -a 256 "$DEST.zip" > "$DEST.zip.sha256"
echo "Created $DEST.zip"
