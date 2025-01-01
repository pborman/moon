// Package moon produces images of the moon at different phases.
//
// This package has builtin images of the moon that are 64x64, 256x256,
// 1024x1025, and 1553x1553.  The source image used is the smallest image that
// is at least big as the size requested. E.g., a 56x56 image will be scaled
// down from the 64x64 image while a 728x728 will be scaled down from the
// 1024x1024 image.
//
// The source image of the moon is from
// https://www.pexels.com/photo/photo-of-full-moon-975012/.
package moon

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"sort"

	"github.com/disintegration/imaging"
	"github.com/llgcode/draw2d/draw2dimg"
)

/*
var (
	Transparent color.RGBA
	White       = color.RGBA{0xff, 0xff, 0xff, 0xff}
	Black       = color.RGBA{0, 0, 0, 0xff}
	Grey        = color.RGBA{0x80, 0x80, 0x80, 0xff}
)
*/

type moonImage struct {
	size  int
	image image.Image
}

var moonImages []moonImage

func register(size int, data []byte) {
	moon, err := png.Decode(bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	moonImages = append(moonImages, moonImage{size: size, image: moon})
	sort.Slice(moonImages, func(i, j int) bool { return moonImages[i].size < moonImages[j].size })
}

// DrawPhaseMask draws a path in gc for the requested phase of the moon for an
// that is width x height.  See Draw for a description of the phase value.
// The mask is always drawn in the box {0,0} {width,height}.
func DrawPhaseMask(gc *draw2dimg.GraphicContext, width, height int, phase float64) {
	cx := float64(width / 2)
	cy := float64(height / 2)
	gc.MoveTo(cx, cy*2)
	phase = -phase
	if phase < 0 {
		phase = -1 - phase
		phase += 0.5
		phase *= 2
		gc.ArcTo(cx, cy, cx, cy, math.Pi/2, math.Pi)
		gc.ArcTo(cx, cy, cx*phase, cy, -math.Pi/2, math.Pi)
	} else {
		phase -= 0.5
		phase *= 2
		gc.ArcTo(cx, cy, cx, cy, math.Pi/2, -math.Pi)
		gc.ArcTo(cx, cy, cx*phase, cy, -math.Pi/2, -math.Pi)
	}
	gc.Close()
}

// Draw returns an image of moon with the provided phase of the given size.
// The shadow is now much illumination the non-visible part should have.
// A shadow of 0 will make it pure black.  A shadow of 1 will not shade it at all.
// The phase is how much of the moon is illuminated.
// A negative phase is a waxing moon.
// A positive phase is a waning moon.
// When waxing the phase goes from 0 to -1.
// When waning the phase goes from 1 to 0
// Full cycle of the moon goes from 0...-1,1...0
func Draw(size int, phase, shadow float64) image.Image {
	if len(moonImages) == 0 {
		return nil
	}
	var mi moonImage
	for _, img := range moonImages {
		mi = img
		if img.size >= size {
			break
		}
	}
	moon := mi.image
	if mi.size != size {
		moon = imaging.Resize(moon, size, size, imaging.Box)
	}
	return DrawFromImage(moon, phase, shadow)
}

// DrawFromImage is like Draw but the caller provides the image of the moon to
// use.  Other than that is is like Draw.
func DrawFromImage(moon image.Image, phase, shadow float64) image.Image {
	mask := image.NewRGBA(moon.Bounds())
	gc := draw2dimg.NewGraphicContext(mask)

	b := moon.Bounds()
	width, height := b.Max.X-b.Min.X, b.Max.Y-b.Min.Y

	w := 1.0
	if phase < 0 {
		w = -1.0
		phase = -phase
	}

	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, uint8(shadow * 255)})
	DrawPhaseMask(gc, width, height, -w*(1-phase))
	gc.Fill()
	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	DrawPhaseMask(gc, width, height, w*phase)
	gc.Fill()

	result := image.NewRGBA(moon.Bounds())
	draw.Draw(result, result.Bounds(), &image.Uniform{color.Transparent}, image.ZP, draw.Src)
	gc = draw2dimg.NewGraphicContext(result)

	gc.SetFillColor(color.Black)
	DrawPhaseMask(gc, width, height, 1)
	gc.Fill()

	draw.DrawMask(result, result.Bounds(), moon, image.Point{0, 0}, mask, image.Point{0, 0}, draw.Over)
	return result
}

// DrawMoonIcon returns an image of size x size with the supplied background and
// a moon drawn as in Draw with he specified shadow and light colors.
func FillMoonIcon(r draw.Image, light, shadow color.Color, phase float64) image.Image {
	gc := draw2dimg.NewGraphicContext(r)
	b := r.Bounds()
	width, height := b.Max.X-b.Min.X, b.Max.Y-b.Min.Y
	gc.SetFillColor(shadow)
	DrawPhaseMask(gc, width, height, 1)
	gc.Fill()
	gc.SetFillColor(light)
	DrawPhaseMask(gc, width, height, phase)
	gc.Fill()
	return r
}
func StrokeMoonIcon(r draw.Image, light, shadow color.Color, phase float64) {
	gc := draw2dimg.NewGraphicContext(r)
	b := r.Bounds()
	width, height := b.Max.X-b.Min.X, b.Max.Y-b.Min.Y
	gc.SetStrokeColor(shadow)
	DrawPhaseMask(gc, width, height, 1)
	gc.Stroke()
	gc.SetStrokeColor(light)
	DrawPhaseMask(gc, width, height, phase)
	gc.Stroke()
}
