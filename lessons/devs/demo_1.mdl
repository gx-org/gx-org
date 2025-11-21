# Compiler errors

```code:main
package main

import "shapes"
import "num"
import "math"

var QKVDim intlen

func rope(x [QKVDim]float32, pos float32) [QKVDim]float32 {
    maxWavelength := [QKVDim/2]float32(10000)
    maxIota := 2 / float32(QKVDim)
    freqExp := [...]float32(num.IotaFull([]intlen{QKVDim/2})) * maxIota
    timescale := math.Pow(maxWavelength, freqExp)
    radians := pos / timescale
    sin, cos := math.Sin(radians), math.Cos(radians)
    xx := shapes.Split(0, x, 2)
    return shapes.Concat(0, 
        xx[0]*cos-xx[1]*sin, 
        xx[1]*cos+xx[0]*sin,
    ).([QKVDim]float32)
}
```

```code:options
{ "hide_text": true }
```
