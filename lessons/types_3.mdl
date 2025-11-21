## Arrays with Symbolic Axis Lengths

Arrays support symbolic axis lengths that will be defined by the host language. A symbolic axis length is declared using `var` at the package level and use the static basic type `intlen`.

Such symbolic axis length, are defined by the host language at runtime. See the config below that defines `Width` and `Height` for GX.
The host language defines these values before compiling the function. For some backend, changing the value of a static variable triggers a recompilation of the function (XLA for example).


```code:config
{
    "Width": 3,
    "Height": 2
}
```

```code:main
package main

import "num"

var Width intlen
var Height intlen

func Main() [Width][Height]int64 {
    return num.IotaFull([]intlen{Width, Height})
}
```
