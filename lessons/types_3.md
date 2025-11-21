## Arrays with Symbolic Axis Lengths

Arrays support symbolic axis lengths that will be defined by the host language. A symbolic axis length is declared using `var` at the package level and use the static basic type `intlen`.

Such symbolic axis length, are defined by the host language at runtime. Here, `Width` and `Height` are defined by the config below.


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
