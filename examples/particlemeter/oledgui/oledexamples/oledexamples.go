/*
oledexamples show what to render
*/
package main

import (
	"fmt"
	"image"

	"oledgui"

	"github.com/hjkoskel/gomonochromebitmap"
)

func main() {

	picrender := gomonochromebitmap.BlockGraphics{
		Clear:       false,
		HaveBorder:  true,
		BorderColor: gomonochromebitmap.FGANSI_YELLOW,
		TextColor:   gomonochromebitmap.FGANSI_BLUE,
	}

	dataX := []float64{-1, 1, 2, 3, 4, 5, 6, 7}
	dataY := []float64{-5, 2, 5.2, 2.4, 5.3, 1.2, -1.5, 0.2}

	pic := gomonochromebitmap.NewMonoBitmap(128, 64, false)

	ploW := oledgui.PlotView{
		Area: image.Rect(12, 3, 100, 60),
		//Range matching to area where plot
		Xmin: -8,
		Xmax: 10,
		Ymin: -6,
		Ymax: 6,
	}
	ploW.AXMarks = ploW.XTicksLin(5, 2.5, true)
	ploW.AYMarks = ploW.YTicksLin(2, 1, true)

	//ploW.AutoscaleX(dataX)
	//ploW.AutoscaleY(dataY)

	fmt.Printf("AXmarks: %s\n", ploW.AXMarks)
	fmt.Printf("AYmarks: %s\n", ploW.AYMarks)

	errPlot := ploW.DrawPlotLine(&pic, dataX, dataY)
	if errPlot != nil {
		fmt.Printf("err plot %s\n", errPlot)
		return
	}
	axStyle := oledgui.PlotAxisStyle{
		AxisWidth:    1,
		TickMajorPos: 3,
		TickMajorNeg: 3,
		TickMinorPos: 1,
		TickMinorNeg: 1,
		ArrowWidth:   2,
		AxisFont:     gomonochromebitmap.GetFont_3x6(),
	}
	ploW.DrawAxis(&pic, axStyle)

	//pic.Rectangle(ploW.Area)
	//drawRect(&pic, ploW.Area)
	fmt.Printf("%s\n", picrender.ToHalfBlockChars(&pic))

	fmt.Printf("KOE: %s\n", ploW.AYMarks)
}
