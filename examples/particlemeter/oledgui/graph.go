/*
graph plotting
*/
package oledgui

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"github.com/hjkoskel/gomonochromebitmap"
)

func roundArr(arr []float64, multiplier float64, offset float64) []int {
	result := make([]int, len(arr))
	for i, f := range arr {
		result[i] = int(math.Round(f*multiplier + offset))
	}
	return result
}

// List last values where it ends. +-inf is not break it is way of plotting. Include +-inf as value
func isNanVec(arr []float64) []bool {
	result := make([]bool, len(arr))
	for i, f := range arr {
		result[i] = math.IsNaN(f)
	}
	return result
}

type AxisMarkerType int

const (
	MAJORTICK AxisMarkerType = 0
	MINORTICK AxisMarkerType = 1
)

func (a AxisMarkerType) String() string {
	s, haz := map[AxisMarkerType]string{
		MAJORTICK: "MAJOR",
		MINORTICK: "MINOR",
	}[a]
	if haz {
		return s
	}
	return "unk"
}

// Axis marker marks something on axis
type AxisMarker struct {
	Type     AxisMarkerType
	Position float64
	Txt      string
}

func (a AxisMarker) String() string {
	return fmt.Sprintf("%s: %s@%.5f", a.Type, a.Txt, a.Position)
}

type AxisMarkers []AxisMarker

func (p *AxisMarkers) GetPositions() []float64 {
	result := make([]float64, len(*p))
	for i, a := range *p {
		result[i] = a.Position
	}
	return result

}

type PlotView struct {
	Area image.Rectangle //Placement and scale on picture or on screen
	//Range matching to area where plot
	Xmin float64
	Xmax float64
	Ymin float64
	Ymax float64

	AXMarks AxisMarkers
	AYMarks AxisMarkers

	//TODO how set log
	LogarithmX float64 //0=linear
	LogarithmY float64
}

func (p *PlotView) AutoscaleX(arr []float64) { //TODO intelligent adjustment
	if len(arr) == 0 {
		return
	}
	p.Xmax = arr[0]
	p.Xmin = arr[0]
	for _, v := range arr {
		p.Xmax = max(p.Xmax, v)
		p.Xmin = min(p.Xmin, v)
	}
}

func (p *PlotView) AutoscaleY(arr []float64) {
	if len(arr) == 0 {
		return
	}
	p.Ymax = arr[0]
	p.Ymin = arr[0]
	for _, v := range arr {
		p.Ymax = max(p.Ymax, v)
		p.Ymin = min(p.Ymin, v)
	}
}

func Tics(typ AxisMarkerType, minvalue float64, maxvalue float64, step float64, labels bool) AxisMarkers {
	if step == 0 {
		return []AxisMarker{}
	}
	nStart := int(math.Ceil(minvalue / step))
	nEnd := int(math.Floor(maxvalue / step))

	q := nEnd - nStart - 1
	//fmt.Printf("minvalue=%v maxvalue=%v step=%v q=%v\n", minvalue, maxvalue, step, q)
	fmtstr := decimalFormatString(step)

	result := make([]AxisMarker, q)
	n := 0
	for i := nStart; i < nEnd; i++ {
		if i == 0 {
			continue
		}
		v := float64(i) * step //TODO log vs linear?

		result[n].Position = v
		result[n].Type = typ
		if labels {
			result[n].Txt = fmt.Sprintf(fmtstr, v)
			//fmt.Printf("LABEL %f  format %s  txt %s\n", v, fmtstr, result[n].Txt)
		}
		n++
	}
	return result
}

// Helper function
func (p *PlotView) XTicksLin(majorTickStep float64, minorTickStep float64, majorLabels bool) AxisMarkers {
	return append(
		Tics(MAJORTICK, p.Xmin, p.Xmax, majorTickStep, majorLabels),
		Tics(MINORTICK, p.Xmin, p.Xmax, minorTickStep, false)...)
}
func (p *PlotView) YTicksLin(majorTickStep float64, minorTickStep float64, majorLabels bool) AxisMarkers {
	return append(
		Tics(MAJORTICK, p.Ymin, p.Ymax, majorTickStep, majorLabels),
		Tics(MINORTICK, p.Ymin, p.Ymax, minorTickStep, false)...)
}

/*
// give values, centered around 0
func tics(minvalue float64, maxvalue float64, step float64) []float64 {
	if step == 0 {
		return []float64{}
	}
	nStart := int(math.Ceil(minvalue / step))
	nEnd := int(math.Floor(maxvalue / step))

	q := nEnd - nStart - 1

	result := make([]float64, q)
	n := 0
	for i := nStart; i < nEnd; i++ {
		if i == 0 {
			continue
		}
		result[n] = float64(i) * step
		n++
	}
	return result
}

func (p *PlotView) XTicks() []float64 {
	return tics(p.Xmin, p.Xmax, p.XTick)
}
func (p *PlotView) YTicks() []float64 {
	return tics(p.Ymin, p.Ymax, p.YTick)
}
*/

func (p *PlotView) ToXPixelScale(arr []float64) []int {
	d := (p.Xmax - p.Xmin)
	ad := float64(p.Area.Dx())
	//xp:= x*area.Dx()/d - area.Dx()*xMin/d  + area.Min.X
	return roundArr(arr, ad/d, -(ad*p.Xmin)/d+float64(p.Area.Min.X))
}
func (p *PlotView) ToYPixelScale(arr []float64) []int {
	d := (p.Ymax - p.Ymin)
	ad := float64(p.Area.Dy())
	return roundArr(arr, -ad/d, -(ad*p.Ymin)/d+float64(p.Area.Min.Y))
}

func (p *PlotView) XMarkersToPixelScale() []int {
	return p.ToXPixelScale(p.AXMarks.GetPositions())
}

func (p *PlotView) YMarkersToPixelScale() []int {
	return p.ToYPixelScale(p.AYMarks.GetPositions())
}

func (p *PlotView) DrawPlotLine(bm *gomonochromebitmap.MonoBitmap, xdata []float64, ydata []float64) error {
	if len(xdata) != len(ydata) {
		return fmt.Errorf("data lens xdata=%v and ydata=%v must be same ", len(xdata), len(ydata))
	}
	if len(xdata) == 0 {
		return nil
	}

	/*
		ratio := (x - xMin) / (xMax - xMin)
		xp := ratio*area.Dx() + area.Min.X
		--------
		d:=xMax-xMin
		ratio:= x/d - xMin/d
		xp := ratio*area.Dx() + area.Min.X
		----

		xp:= x*area.Dx()/d - area.Dx()*xMin/d  + area.Min.X

	*/
	intXArr := p.ToXPixelScale(xdata)
	intYArr := p.ToYPixelScale(ydata)

	nanX := isNanVec(xdata)
	nanY := isNanVec(ydata)

	//fmt.Printf("x:%#v\ny:%#v\n", intXArr, intYArr)
	//Clamp points

	//Fact:128 is not much pixels  Clamp to 128 long arr and then render
	havePrev := false
	for i, xv := range intXArr {
		if nanX[i] || nanY[i] {
			havePrev = false
			continue
		}

		if havePrev { //have previous point where draw line
			yv := intYArr[i]
			p0 := image.Point{X: intXArr[i-1], Y: intYArr[i-1]}
			p1 := image.Point{X: xv, Y: yv}

			p0Clip, p1Clip := gomonochromebitmap.ClipLine(p0, p1, p.Area)
			if p0Clip != nil && p1Clip != nil {
				//fmt.Printf("line %s to %s\n", p0Clip, p1Clip)
				bm.Line(*p0Clip, *p1Clip, true)
			}
		}
		havePrev = true
	}
	return nil
}

type PlotAxisStyle struct {
	AxisWidth    int
	TickMajorPos int //Pixels up from axis
	TickMajorNeg int //Pixels down from axis
	TickMinorPos int //Pixels up from axis
	TickMinorNeg int //Pixels down from axis

	ArrowWidth int

	AxisFont map[rune]gomonochromebitmap.MonoBitmap
}

func decimalFormatString(step float64) string {
	d, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", step-math.Trunc(step)), 64) //TODO solve rounding strangeness
	if d < 0.0000001 {
		return "%.0f"
	}
	//fmt.Printf("step=%.5f d=%.6f log10=%.4f\n", step, d, -math.Log10(d))
	n := max(0, math.Ceil(-math.Log10(d)))
	//fmt.Printf("step=%v n=%v\n", step, n)
	return fmt.Sprintf("%%.%vf", n)
}

func (p *PlotView) DrawAxis(bm *gomonochromebitmap.MonoBitmap, style PlotAxisStyle) error {
	xorig := p.ToXPixelScale([]float64{0})[0] //Stupid? optimize?
	yorig := p.ToYPixelScale([]float64{0})[0]

	if p.Area.Min.X <= xorig && xorig <= p.Area.Max.X {
		bm.Vline(xorig, p.Area.Min.Y, p.Area.Max.Y, true)
	}

	if p.Area.Min.Y <= yorig && yorig <= p.Area.Max.Y {
		bm.Hline(p.Area.Min.X, p.Area.Max.X, yorig, true)
	}

	//----------------------------
	zerochar := style.AxisFont['0']
	markXPos := p.XMarkersToPixelScale()
	for i, mark := range p.AXMarks {
		x := markXPos[i]
		wTxt := zerochar.W * len(mark.Txt)
		hTxt := zerochar.H
		switch mark.Type {
		case MAJORTICK:
			y0 := yorig - style.TickMajorNeg
			y1 := yorig + style.TickMajorPos
			bm.Vline(x, y0, y1, true)
			bm.Print(mark.Txt, style.AxisFont, 0, 1, image.Rect(x-wTxt/2, y1+2, x, y1+hTxt+2), true, true, false, false)
		case MINORTICK:
			y0 := yorig - style.TickMinorNeg
			y1 := yorig + style.TickMinorPos
			bm.Vline(x, y0, y1, true)
			bm.Print(mark.Txt, style.AxisFont, 0, 1, image.Rect(x-wTxt/2, y1+2, x, y1+hTxt+2), true, true, false, false)
		}
	}

	markYPos := p.YMarkersToPixelScale()
	for i, mark := range p.AYMarks {
		y := markYPos[i]
		wTxt := zerochar.W * len(mark.Txt)
		hTxt := zerochar.H
		switch mark.Type {
		case MAJORTICK:
			x0 := xorig - style.TickMajorNeg
			x1 := xorig + style.TickMajorPos
			bm.Hline(x0, x1, y, true)
			bm.Print(mark.Txt, style.AxisFont, 0, 1, image.Rect(x0-wTxt-2, y-hTxt/2, x0, y+hTxt*2), true, false, false, false)
		case MINORTICK:
			x0 := xorig - style.TickMinorNeg
			x1 := xorig + style.TickMinorPos
			bm.Hline(x0, x1, y, true)
			bm.Print(mark.Txt, style.AxisFont, 0, 1, image.Rect(x0-wTxt-2, y-hTxt/2, x0, y+hTxt*2), true, true, false, false)
		}
	}

	//arrow
	for i := 0; i <= style.ArrowWidth; i++ {
		bm.SetPix(p.Area.Max.X-i, yorig-i, true)
		bm.SetPix(p.Area.Max.X-i, yorig+i, true)

		bm.SetPix(xorig-i, p.Area.Min.Y+i, true)
		bm.SetPix(xorig+i, p.Area.Min.Y+i, true)
	}

	return nil
}
