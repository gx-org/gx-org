package lessons

import "embed"

//go:embed *.mdl devs/*.mdl demos/*.mdl
var Lessons embed.FS
