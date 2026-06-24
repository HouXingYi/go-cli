#!/usr/bin/env bash
set -euo pipefail

APP_NAME="zn-cli"
ENTRY="./cmd/ziniao"
DIST_DIR="dist"

targets=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
)

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

for target in "${targets[@]}"; do
  read -r goos goarch <<< "$target"

  output="$DIST_DIR/$APP_NAME-$goos-$goarch"
  if [[ "$goos" == "windows" ]]; then
    output="$output.exe"
  fi

  echo "Building $output"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -o "$output" "$ENTRY"
done

echo "Build artifacts are in $DIST_DIR/"
