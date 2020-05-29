package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
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

var palette []uint32
var lumaCache map[uint32]uint64 = make(map[uint32]uint64)
var paletteSet map[uint32]bool

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
	/*	fmt.Println("Reading image...")
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
		}*/

	fmt.Println("Generating OC palette...")
	palette, _ = generatePalette()

	fmt.Println("Calculating color importances...")
	for i := 0; i < len(palette); i++ {
		lumaCache[palette[i]] = CalculateColorLuma(palette[i])
	}

	fmt.Println("Downsampling video...")
	os.RemoveAll("tmp/")
	os.MkdirAll("./tmp/", 0755)
	cmd := exec.Command("ffmpeg", "-i", os.Args[1], "-vf", "fps=10,scale=320:200", "tmp/out%06d.png")
	cmd.Run()

	fmt.Println("Calculating diffs...")

	files, _ := ioutil.ReadDir("tmp")

	os.RemoveAll("diffs/")
	os.MkdirAll("./diffs/", 0755)

	a, _ := ioutil.ReadFile("tmp/out000001.png")
	fd, _ := os.Create("diffs/out000001.png")
	fd.Write(a)
	fd.Close()

	for i := 1; i < len(files); i++ {
		from, _ := os.Open("tmp/" + files[i-1].Name())
		to, _ := os.Open("tmp/" + files[i].Name())
		defer from.Close()
		defer to.Close()
		fmt.Println(files[i-1].Name(), "->", files[i].Name())

		fromI, _, _ := image.Decode(from)
		toI, _, _ := image.Decode(to)

		u := ImageDiff(fromI, toI)
		fd, _ := os.Create("diffs/" + files[i].Name())
		png.Encode(fd, u)
		defer fd.Close()
	}

	os.RemoveAll("tmpfif")
	os.Mkdir("tmpfif", 0755)

	for i := 0; i < len(files); i++ {
		in, _ := os.Open("diffs/" + files[i].Name())
		defer in.Close()

		inimage, _, _ := image.Decode(in)
		fifData := encodeFif(320, 200, inimage, false)
		fd, _ := os.Create("tmpfif/" + files[i].Name() + ".fif")
		fd.Write(fifData)
		fd.Close()
	}

	// Write the MASTERFIF™
	fif := []byte("FastIF")
	fif = append(fif, 160, 50)

	for i := 0; i < len(files); i++ {
		in, _ := ioutil.ReadFile("tmpfif/" + files[i].Name() + ".fif")
		in = in[8:]
		in = in[:len(in)-1]
		fif = append(fif, in...)
		fif = append(fif, 0x12, 5)
	}

	fif = append(fif, 0x20)
	af, _ := os.Create(os.Args[2])
	af.Write(fif)
	af.Close()
	/*
		a := time.Now()
		data := encodeFif(w, h, src)
		fmt.Println(time.Since(a))
		fd, _ := os.Create(os.Args[2])
		fd.Write(data)
		fd.Close()*/
}
