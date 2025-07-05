package lessons

import "embed"

//go:embed *.md devs/*.md
var Lessons embed.FS
