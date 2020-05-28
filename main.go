package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"time"
)

func splitColor(color uint32) (uint32, uint32, uint32) {
	return color >> 16 & 0xFF, color >> 8 & 0xFF, color & 0xFF
}

func colorDelta(c uint32, d uint32) int64 {
	r, g, b := splitColor(c)
	tR, tG, tB := splitColor(d)

	factorR := int64(int(tR) - int(r))
	factorG := int64(int(tG) - int(g))
	factorB := int64(int(tB) - int(b))
	delta := (factorR * factorR) + (factorG * factorG) + (factorB * factorB)

	return delta
}

var palette        []      uint32
var lumaCache   map[uint32]uint64 = make(map[uint32]uint64)
var paletteSet  map[uint32]bool

func CalculateColorLuma(col uint32) uint64 {
	rr, gg, bb := splitColor(col)
	r, g, b := float64(rr), float64(gg), float64(bb)

	r = math.Pow(r, 2)
	g = math.Pow(g, 2)
	b = math.Pow(b, 2)

	r = .299 * r
	g = .587 * g
	b = .114 * b

	importance := math.Sqrt(r + g + b)
	importance = importance * 1000

	return uint64(importance) + 1000
}

var FIF_altMode = true

func main() {
	fmt.Println("Reading image...")
	in, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("An error has occured when reading the image.", err)
		panic("Program cannot continue.")
	}
	defer in.Close()

	src, _, _ := image.Decode(in)
	bounds := src.Bounds()
	w := bounds.Size().X
	h := bounds.Size().Y
	if w > 320 || h > 200 {
		fmt.Println("Image too sizeable.")
		panic("Program cannot continue")
	}

	if w % 2 != 0 || h % 4 != 0 {
		panic("w must be a multiple of 2 and h must be a multiple of 4")
	}


	fmt.Println("Generating OC palette...")
	palette, _ = generatePalette()

	fmt.Println("Calculating color importances...")
	for i := 0; i < len(palette); i++ {
		lumaCache[palette[i]] = CalculateColorLuma(palette[i])
	}

	a := time.Now()
	data := encodeFif(w, h, src)
	fmt.Println(time.Since(a))
	fd, _ := os.Create(os.Args[2])
	fd.Write(data)
	fd.Close()
}
