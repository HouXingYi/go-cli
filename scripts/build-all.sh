#!/usr/bin/env bash
set -euo pipefail

DIST_DIR="dist"

variants=(
  "zn-ent ./cmd/zn-ent"
  "zn-eco ./cmd/zn-eco"
)

targets=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
  "windows amd64"
)

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

for variant in "${variants[@]}"; do
  read -r app_name entry <<< "$variant"

  for target in "${targets[@]}"; do
    read -r goos goarch <<< "$target"

    output="$DIST_DIR/$app_name-$goos-$goarch"
    if [[ "$goos" == "windows" ]]; then
      output="$output.exe"
    fi

    echo "Building $output"
    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -o "$output" "$entry"
  done
done

echo "Build artifacts are in $DIST_DIR/"
