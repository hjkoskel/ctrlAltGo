package puhveri

import (
	"image"
	"image/color"
)

// 16 bit
type ImageBGR565 struct {
	Pix    []byte
	Rect   image.Rectangle
	Stride int
}

func (i *ImageBGR565) GetRaw() []byte {
	return i.Pix
}

func (i *ImageBGR565) Bounds() image.Rectangle { return i.Rect }
func (i *ImageBGR565) ColorModel() color.Model { return color.NRGBAModel }

func (i *ImageBGR565) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(i.Rect)) {
		return color.NRGBA{}
	}

	pix := i.Pix[i.PixOffset(x, y):]
	return color.NRGBA{
		R: (pix[1] >> 3) << 3,
		G: (pix[1] << 5) | ((pix[0] >> 5) << 2),
		B: pix[0] << 3,
		A: 255,
	}
}

func (i *ImageBGR565) Set(x, y int, c color.Color) {
	i.SetNRGBA(x, y, color.NRGBAModel.Convert(c).(color.NRGBA))
}

func (i *ImageBGR565) SetNRGBA(x, y int, c color.NRGBA) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}

	pix := i.Pix[i.PixOffset(x, y):]
	pix[0] = (c.B >> 3) | ((c.G >> 2) << 5)
	pix[1] = (c.G >> 5) | ((c.R >> 3) << 3)
}

func (i *ImageBGR565) PixOffset(x, y int) int {
	return (y-i.Rect.Min.Y)*i.Stride + (x-i.Rect.Min.X)*2
}

func (i *ImageBGR565) Fill(area image.Rectangle, c color.Color) {

}
