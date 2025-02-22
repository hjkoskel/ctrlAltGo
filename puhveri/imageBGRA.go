package puhveri

import (
	"image"
	"image/color"
)

// 32bit colors but in wrong order :)
type ImageBGRA struct {
	Pix    []byte
	Rect   image.Rectangle
	Stride int
}

func (i *ImageBGRA) GetRaw() []byte {
	return i.Pix
}

func (i *ImageBGRA) Bounds() image.Rectangle { return i.Rect }
func (i *ImageBGRA) ColorModel() color.Model { return color.RGBAModel }

func (i *ImageBGRA) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(i.Rect)) {
		return color.RGBA{}
	}

	pix := i.Pix[i.PixOffset(x, y):]
	return color.RGBA{
		pix[2],
		pix[1],
		pix[0],
		pix[3],
	}
}

func (i *ImageBGRA) Set(x, y int, c color.Color) {
	i.SetRGBA(x, y, color.RGBAModel.Convert(c).(color.RGBA))
}

func (i *ImageBGRA) SetRGBA(x, y int, c color.RGBA) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}

	n := i.PixOffset(x, y)
	pix := i.Pix[n:]
	pix[0] = c.B
	pix[1] = c.G
	pix[2] = c.R
	pix[3] = c.A
}

func (i *ImageBGRA) PixOffset(x, y int) int {
	return (y-i.Rect.Min.Y)*i.Stride + (x-i.Rect.Min.X)*4
}

func (i *ImageBGRA) Fill(area image.Rectangle, c color.Color) {
	are := area.Intersect(i.Bounds())
	cc := color.RGBAModel.Convert(c).(color.RGBA)
	rowStart := i.PixOffset(are.Min.X, are.Min.Y)
	dx := are.Dx()
	for y := are.Min.Y; y < are.Max.Y; y++ {
		pix := i.Pix[rowStart:]
		index := 0
		for x := 0; x < dx; x++ {
			pix[index] = cc.B
			pix[index+1] = cc.G
			pix[index+2] = cc.R
			pix[index+3] = cc.A
			index += 4
		}
		rowStart += i.Stride
	}
}
