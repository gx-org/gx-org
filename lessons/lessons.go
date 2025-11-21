package lessons

import "embed"

//go:embed *.mdl devs/*.mdl
var Lessons embed.FS
