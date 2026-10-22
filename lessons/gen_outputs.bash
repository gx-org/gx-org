#!/bin/bash

cd "$(dirname "$0")"

# Remove all existing files.
rm -v *.gx.out

# Format all GX files.
go tool gx fmt -w .

# Runs all GX lessons and stores their output in <name>.gx.out.
for f in *.gx */*.gx; do
  go tool gx run "$f" > "$f.out"
done
