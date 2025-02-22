/*
Simple graphics demo

no networking, just minimalistic framebuffer demo

runs as PID1 and also by loaded by programManager
*/
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"puhveri"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/math/fixed"
)

const (
	BLOCKW = 32
	BLOCKH = 32
)

type BounceBlock struct {
	X  float64
	Y  float64
	Dx float64 //pixels per ms
	Dy float64

	Width  int
	Heigth int

	C color.RGBA
}

func (p *BounceBlock) Rect() image.Rectangle {
	return image.Rect(int(p.X), int(p.Y), int(p.X)+p.Width, int(p.Y)+p.Heigth)
}

func (p *BounceBlock) Step(tDelta time.Duration, area image.Rectangle) {
	xNow := p.X + p.Dx*float64(tDelta.Microseconds())/1000
	yNow := p.Y + p.Dy*float64(tDelta.Microseconds())/1000

	if int(xNow) < area.Min.X || area.Max.X < int(xNow)+p.Width {
		p.Dx *= -1
		xNow = p.X + p.Dx*float64(tDelta.Microseconds())/1000
	}

	if int(yNow) < area.Min.Y || area.Max.Y < int(yNow)+p.Heigth {
		p.Dy *= -1
		yNow = p.Y + p.Dy*float64(tDelta.Microseconds())/1000
	}
	p.X = xNow
	p.Y = yNow
}

func main() {

	if os.Getpid() == 1 {
		//Need to initialize stuff, acts as init system
		errMount := initializing.MountNormal()
		ctrlaltgo.JamIfErr(errMount)
	}

	lastFbDevName, errName := puhveri.GetLastFbFileName()
	if errName != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("failed getting FB name %s", errName))
	}

	dev, devErr := puhveri.OpenFbPuhveri(lastFbDevName)
	if devErr != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("opening fb device failed %s", devErr))
	}

	errConsole := puhveri.ConsoleToGraphics()
	if errConsole != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("error chancing to console: %s", errConsole))
	}

	imgBuffer, errImgBuffer := dev.GetBuffer()
	if errImgBuffer != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("error getting framebuffer:%s", errImgBuffer))
	}

	//Rectangle fill demo point out that video mode change works
	bou := imgBuffer.Bounds()

	blocks := []BounceBlock{
		BounceBlock{
			X:     float64(bou.Dx()) / 2,
			Y:     float64(bou.Dy()) / 2,
			Dx:    0.15,
			Dy:    0.32,
			Width: 10, Heigth: 10,
			C: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		BounceBlock{
			X:     float64(bou.Dx()) / 2,
			Y:     float64(bou.Dy()) / 2,
			Dx:    -0.15,
			Dy:    -0.32,
			Width: 20, Heigth: 20,
			C: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
		BounceBlock{
			X:     float64(bou.Dx()) / 2,
			Y:     float64(bou.Dy()) / 2,
			Dx:    -0.32,
			Dy:    0.05,
			Width: 40, Heigth: 10,
			C: color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
	}

	//Get font
	monofont, errMono := truetype.Parse(gomono.TTF)
	if errMono != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("error getting truetype font %s\n", errMono))
	}
	monoface := truetype.NewFace(monofont, &truetype.Options{Size: 16})

	tPrevSim := time.Now()

	textColorPic := image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255}) //REally stupid way how this is done whith golang image system
	for {
		//Clear
		imgBuffer.Fill(bou, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		dur := time.Since(tPrevSim)
		for i, _ := range blocks {
			blocks[i].Step(dur, bou)
			imgBuffer.Fill(blocks[i].Rect(), blocks[i].C)
		}
		d := &font.Drawer{
			Dst:  imgBuffer,
			Src:  textColorPic,
			Face: monoface,
			Dot:  fixed.P(2, 32),
		}

		d.DrawString(fmt.Sprintf("fps:%.2f", 1/float64(dur.Seconds())))

		tPrevSim = time.Now()
		dev.UpdateScreenAreas(imgBuffer, []image.Rectangle{})
	}
}
