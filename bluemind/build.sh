#!/bin/bash
# Builds the nfpm binaries published as net.bluemind:nfpm (see pom.xml).
set -euo pipefail

cd "$(dirname "$0")/.."
VERSION=$(sed -n 's:.*<version>\(.*\)</version>.*:\1:p' bluemind/pom.xml | head -1)
DIST=bluemind/target/dist
rm -rf "$DIST"
mkdir -p "$DIST"

for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
    os=${platform%/*}
    arch=${platform#*/}
    work="$DIST/$os-$arch"
    mkdir -p "$work"
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath \
        -ldflags "-s -w -X main.version=$VERSION -X main.commit=$(git rev-parse --short HEAD) -X main.builtBy=bluemind" \
        -o "$work/nfpm" ./cmd/nfpm
    tar -czf "$DIST/nfpm-$VERSION-$os-$arch.tar.gz" -C "$work" nfpm
    rm -r "$work"
done
ls -l "$DIST"
