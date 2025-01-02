// Package moon produces images of the moon at different phases and provides
// information about the moon's location.
//
// The phase of the moon is represented by a floating point number between -1.0
// and 1.0 where the absolute value is the percentage of the moon that is
// visible.  Both -1. and 1.0 represent a full moon.  Negative numbers indicate
// a waxing moon and positive numbers indicate a waning moon.  A full cycle from
// new moon to new moon is uses the values 0.0 - -1.0 as it waxes and 1.0 - 0.0
// as it wanes.
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
	"time"

	"github.com/disintegration/imaging"
	sampa "github.com/hablullah/go-sampa"
	"github.com/llgcode/draw2d/draw2dimg"
)

type moonImage struct {
	size  int
	image image.Image
}

var moonImages []moonImage

// register is called during init by the various moon*.go files which contain
// PNG versions of the moon in different resolutions.
func register(size int, data []byte) {
	moon, err := png.Decode(bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	moonImages = append(moonImages, moonImage{size: size, image: moon})
	sort.Slice(moonImages, func(i, j int) bool { return moonImages[i].size < moonImages[j].size })
}

// DrawPhaseMask draws a path in to gc for the requested phase of the moon that
// has the supplied width and height.  Typically these are the same value.  The
// mask is always drawn in the box {0,0} with the center of the moon at width/2,
// height/2.  The calling function can then call gc.Stroke() or gc.Fill() to
// draw or fill in the path.
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

// FillMoonIcon draws a 2 color image of the moon with the illuminated portion
// being drawing in light and the shaded portion drawn in shado.
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

// StrokMeoonIcon is similar to DrawMoonIcon but only draws the outline.  A full
// circle is always drawn first using the color shadow and the illuminated
// portions of the moon is drawn with the light color.
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

// Information provides information about the moon.  The illumination is a value
// between -1.0 and 1.0.  The amount of illumination is the absolute value
// (e.g., -0.25 and 0.25 both indicate 25% illumination).  If Illumination is
// negative then the moon is waxing.  If positive the moon is waning.  An
// illumination of 0.0 is a new moon and either -1.0 or 1.0 is a full moon.
type Information struct {
	Moonrise     time.Time       // When the moon will/did rise
	Moonset      time.Time       // When the moon will/did set
	Direction    float64         // Direction to the moon
	Elevation    float64         // Elevation of the moon
	Illumination float64         // How illuminated the moon is
	Phase        sampa.MoonPhase // The current phase
}

// Info returns information about the moon at a specified location and time.
func Info(when time.Time, loc sampa.Location) (*Information, error) {
	ev, err := sampa.GetMoonEvents(when, loc, nil)
	if err != nil {
		return nil, err
	}
	p, err := sampa.GetMoonPosition(when, loc, nil)
	if err != nil {
		return nil, err
	}

	info := &Information{
		Elevation:    p.TopocentricElevationAngle,
		Direction:    p.ObserverLocalHourAngle + 180,
		Moonrise:     ev.Moonrise.DateTime,
		Moonset:      ev.Moonset.DateTime,
		Illumination: p.PercentIlluminated,
		Phase:        p.Phase,
	}
	if info.Direction >= 360 {
		info.Direction -= 360
	}
	switch info.Phase {
	case sampa.NewMoon, sampa.WaxingCrescent, sampa.FirstQuarter, sampa.WaxingGibbous:
		info.Illumination = -info.Illumination
	}
	return info, nil
}
