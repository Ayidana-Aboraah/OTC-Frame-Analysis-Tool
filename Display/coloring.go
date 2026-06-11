package Display

import (
	"errors"
	"image"
	"image/color"
	"os"

	"golang.org/x/image/bmp"
)

// Iron-inspired thermal palette.
// Interpolates between these color stops.
var ironPalette = []color.RGBA{
	{0, 0, 0, 255},       // black
	{40, 0, 80, 255},     // dark purple
	{120, 0, 120, 255},   // purple
	{180, 0, 60, 255},    // red-purple
	{220, 30, 0, 255},    // red
	{255, 80, 0, 255},    // orange-red
	{255, 160, 0, 255},   // orange
	{255, 220, 0, 255},   // yellow
	{255, 255, 255, 255}, // white
}

func lerp(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t)
}

func paletteColor(v float64) color.RGBA {
	if v <= 0 {
		return ironPalette[0]
	}
	if v >= 1 {
		return ironPalette[len(ironPalette)-1]
	}

	segments := len(ironPalette) - 1
	pos := v * float64(segments)

	idx := int(pos)
	t := pos - float64(idx)

	c1 := ironPalette[idx]
	c2 := ironPalette[idx+1]

	return color.RGBA{
		R: lerp(c1.R, c2.R, t),
		G: lerp(c1.G, c2.G, t),
		B: lerp(c1.B, c2.B, t),
		A: 255,
	}
}

func TemperaturesToBMP(
	temps []float32,
	width, height int,
	outputPath string,
) error {

	if len(temps) < width*height {
		return errors.New("temperature array size does not match width*height")
	}

	// Find min/max temperature.
	minT := temps[0]
	maxT := temps[0]

	for _, t := range temps[1:] {
		if t < minT {
			minT = t
		}
		if t > maxT {
			maxT = t
		}
	}

	// Avoid divide-by-zero if all temps are equal.
	rangeT := maxT - minT
	if rangeT == 0 {
		rangeT = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x

			normalized := float64(
				(temps[idx] - minT) / rangeT,
			)

			img.SetRGBA(x, y, paletteColor(normalized))
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return bmp.Encode(f, img)
}

func TemperaturesIntToBMP(temps []uint16,
	width, height int,
	outputPath string) error {
	if len(temps) < width*height {
		return errors.New("temperature array size does not match width*height")
	}

	// Find min/max temperature.
	minT := temps[0]
	maxT := temps[0]

	for _, t := range temps[1:] {
		if t < minT {
			minT = t
		}
		if t > maxT {
			maxT = t
		}
	}

	// Avoid divide-by-zero if all temps are equal.
	rangeT := maxT - minT
	if rangeT == 0 {
		rangeT = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x

			normalized := float64((temps[idx] - minT)) / float64(rangeT)

			img.SetRGBA(x, y, paletteColor(normalized))
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return bmp.Encode(f, img)
}
