## Generic Types

Generic functions can be defined for multiple types as long as these types satisfy some constraints. The package [`dtype`](https://github.com/gx-org/gx/blob/main/stdlib/dtype/dtype.gx) defines sets of basic types, notably `Floats`, `Ints`, and `Num` respectively for floats, integers, and numerical types.

Whenever possible, the compiler will use arguments to infer the type parameters of the function, as in the second call to `ToInts` in the example.

```code:main
package main

import "num"
import "dtype"

func ToInts[U dtype.Ints, T dtype.Floats](a [2][3]T) [2][3]U {
    return [2][3]U(a)
}

func Main() ([2][3]int64, [2][3]int64) {
    a := [2][3]float32{{1, 2, 3}, {4, 5, 6}}
    i64a := ToInts[float32, int64](a)
    // In the following call, the compiler uses type inference
    // to determine the type of the argument.
    i64b := ToInts[int64](a)
    return i64a, i64b
}
```
