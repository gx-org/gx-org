#!/bin/bash

cat > buildtime.go << EOF
package buildtime

// BuildTime at which the WASM file has been generated.
const BuildTime = "$(date '+%Y-%m-%d %H:%M:%S')"
EOF
