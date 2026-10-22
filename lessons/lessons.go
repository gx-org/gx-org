package lessons

import "embed"

//go:embed *.mdl *.gx devs/*.mdl devs/*.gx demos/*.mdl demos/*.gx
//go:embed *.gx.out devs/*.gx.out demos/*.gx.out
var Lessons embed.FS
