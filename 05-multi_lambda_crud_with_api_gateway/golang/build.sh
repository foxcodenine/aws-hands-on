#!/usr/bin/env bash
#
# Builds every Lambda in cmd/ for the provided.al2023 runtime.
#
# Each function needs its own binary named "bootstrap", so they go into
# build/<function>/bootstrap - one directory each, named after the cmd folder.
# Terraform then picks them up with for_each over that same list of names,
# instead of a hand-written block per function.
#
# Usage: ./build.sh

set -euo pipefail

cd "$(dirname "$0")"

rm -rf build

for dir in cmd/*/; do
	name=$(basename "$dir")

	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -o "build/$name/bootstrap" "./$dir"

	echo "built build/$name/bootstrap"
done
