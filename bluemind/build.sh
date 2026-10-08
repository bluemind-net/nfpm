#!/bin/bash
set -euo pipefail

VERSION=${1:?usage: bluemind/build.sh <version>}
cd "$(dirname "$0")/.."
DIST=bluemind/target/dist
rm -rf "$DIST"
mkdir -p "$DIST/linux-amd64"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
    -ldflags "-s -w -X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.builtBy=bluemind" \
    -o "$DIST/linux-amd64/nfpm" ./cmd/nfpm
tar -czf "$DIST/nfpm-$VERSION-linux-amd64.tar.gz" -C "$DIST/linux-amd64" nfpm
rm -r "$DIST/linux-amd64"
ls -l "$DIST"
