package lessons

import "embed"

//go:embed *.mdl *.gx devs/*.mdl devs/*.gx demos/*.mdl demos/*.gx
var Lessons embed.FS
