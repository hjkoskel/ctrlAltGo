package puhveri

import (
	"image"
	"image/color"
)

type PuhveriImage interface { //Implements image.Image and draw.Image
	// ColorModel returns the Image's color model.
	ColorModel() color.Model
	// Bounds returns the domain for which At can return non-zero color.
	// The bounds do not necessarily contain the point (0, 0).
	Bounds() image.Rectangle
	// At returns the color of the pixel at (x, y).
	// At(Bounds().Min.X, Bounds().Min.Y) returns the upper-left pixel of the grid.
	// At(Bounds().Max.X-1, Bounds().Max.Y-1) returns the lower-right one.
	At(x, y int) color.Color
	Set(x, y int, c color.Color)
	GetRaw() []byte

	//Few specialized functions for better performance (do color conversion once etc..)
	Fill(area image.Rectangle, c color.Color)
}
