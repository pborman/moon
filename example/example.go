package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/pborman/colors/web"
	"github.com/pborman/moon"
)

func makeIcon(path string, size int, phase float64) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Draw a moon with the visible part white onto img.
	moon.FillMoonIcon(img, web.White, web.Black, phase)
	// Draw the outline of a full moon in black onto img.
	moon.StrokeMoonIcon(img, web.Black, web.Transparent, 1)

	fd, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	png.Encode(fd, img)
	fd.Close()
}

func makePhoto(path string, size int, phase float64) {
	// Get a photo of the moon with the no-visible part being 33% illuminated.
	img := moon.Draw(size, phase, .33)
	fd, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	png.Encode(fd, img)
	fd.Close()
}

func main() {
	makeIcon("moon-new.png", 128, 0)
	makeIcon("moon-waxing-crescent.png", 128, -0.25)
	makeIcon("moon-waning-gibbous.png", 128, 0.75)
	makeIcon("moon-full.png", 128, 1)

	makePhoto("moon-photo-new.png", 128, 0)
	makePhoto("moon-photo-waxing-crescent.png", 128, -0.25)
	makePhoto("moon-photo-waning-gibbous.png", 128, 0.75)
	makePhoto("moon-photo-full.png", 128, 1)
}
