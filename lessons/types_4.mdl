## Static Cast 

For a static cast checked by the compiler, use `T(value)` where `T` is the target type as in `int32(2)`.
That operator can also be used as a reshape operator or to change the data type of an array as demonstrated in the example.

The cast operator can also be used to fill an array with a constant value: for example `[2][3]float32(3)`.

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

func Main() ([Height][Width]float32, [2][3]int32) {
    a := num.IotaFull([]intlen{Width, Height})
    // Reshape from [Width][Height] to [Height][Width] and to float32
    b := [Height][Width]float32(a)
    // Cast and expand the scalar 6 to a [2][3]int32 array.
    // Note the use of () instead of {}.
    c := [2][3]int32(6)
    return b, c
}
```
