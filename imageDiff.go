package main

import (
	"image"
	"image/color"
)

type ImageSegment struct {
	r [8]uint32
	g [8]uint32
	b [8]uint32
	a [8]uint32
}

func (a *ImageSegment) compare(z *ImageSegment) bool {
	for i := 0; i < 8; i++ {
		if a.r[i] != z.r[i] {
			return false
		}
		if a.g[i] != z.g[i] {
			return false
		}
		if a.b[i] != z.b[i] {
			return false
		}
	}

	return true
}

func (a *ImageSegment) write(b *image.RGBA, x int, y int) {
	for i := 0; i < 8; i++ {
		xx := i / 4
		yy := i % 4

		col := color.RGBA{byte(a.r[i]), byte(a.g[i]), byte(a.b[i]), 255}

		b.Set((x*2)+xx, (y*4)+yy, col)
	}
}

func ISFromImage(a image.Image, x int, y int) *ImageSegment {
	is := new(ImageSegment)
	for xx := 0; xx < 2; xx++ {
		for yy := 0; yy < 4; yy++ {
			r, g, b, _ := a.At((x*2)+xx, (y*4)+yy).RGBA()
			r = r / 257
			g = g / 257
			b = b / 257

			zcolor := ((r & 0xFF) << 16) + ((g & 0xFF) << 8) + b&0xFF
			zcolor = simplifyColor(palette, zcolor)

			r, g, b = splitColor(zcolor)

			is.r[(xx*4)+yy] = r
			is.g[(xx*4)+yy] = g
			is.b[(xx*4)+yy] = b
		}
	}

	return is
}

func ImageDiff(from image.Image, to image.Image) image.Image {
	upLeft := image.Point{}
	lowRight := from.Bounds().Size()
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// TODO: Replace with non-hardcoded resolution variables.
	for x := 0; x < 160; x++ {
		for y := 0; y < 50; y++ {
			sr := ISFromImage(from, x, y)
			ds := ISFromImage(to, x, y)
			if sr.compare(ds) != true {
				ds.write(img, x, y)
			}
		}
	}

	return img
}
