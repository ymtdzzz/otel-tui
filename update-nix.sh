#!/usr/bin/env bash

# This script is used to update the nix file with the latest version of otel-tui.
# It is based on joerdav/xc. See https://github.com/joerdav/xc
sed -e "s/__VERSION__/$(git describe --tags --abbrev=0)/g" otel-tui.nix.tmpl > otel-tui.nix

# We first try to build and it fails with hash mismatch, and we use it to populate sha256.
echo "Calculating source sha256..."
nix-build -E 'with import <nixpkgs> { }; callPackage ./otel-tui.nix { }'
SRC_SHA256="$(nix-build -E 'with import <nixpkgs> { }; callPackage ./otel-tui.nix { }' 2>&1 | grep -oE 'got:\s+sha256-[a-zA-Z0-9+/=]+' | cut -d "-" -f 2)"
sed -i -e "s|hash = \"sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\";|hash = \"sha256-$SRC_SHA256\";|g" otel-tui.nix

# We try again to build and it fails with hash mismatch, and we use it to populate vendorSha256.
echo "Calculating vendor sha256..."
VENDOR_SHA256="$(nix-build -E 'with import <nixpkgs> { }; callPackage ./otel-tui.nix { }' 2>&1 | grep -oE 'got:\s+sha256-[a-zA-Z0-9+/=]+' | cut -d "-" -f 2)"
sed -i -e "s|vendorHash = \"sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\";|vendorHash = \"sha256-$VENDOR_SHA256\";|g" otel-tui.nix

rm otel-tui.nix-e
