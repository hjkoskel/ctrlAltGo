/*
gui for generating view for particle meter
*/
package oledgui

import (
	"fmt"
	"strings"
	"time"

	"github.com/hjkoskel/gomonochromebitmap"
)

func ViewMetrics(pm10 float64, pm25 float64, haveAmbient bool, temperature float64, pressure float64, rh float64) gomonochromebitmap.MonoBitmap {
	result := gomonochromebitmap.NewMonoBitmap(128, 64, false)
	titlefont := gomonochromebitmap.GetFont_8x8()

	tNow := time.Now()

	monthLetters := strings.ToLower(tNow.Month().String())[0:3]

	txtTitle := fmt.Sprintf("%02v:%02v:%02v %s%02d", tNow.Hour(), tNow.Minute(), tNow.Second(), monthLetters, tNow.Day())

	b := result.Bounds()
	result.Print(txtTitle,
		titlefont, 0, 0, b, true, true, true, false)

	largefont := gomonochromebitmap.GetFont_11x16()

	b.Min.Y = 12
	txtConcentrations := fmt.Sprintf("Small:%.1f\nLarge:%.1f\n", pm10, pm25)
	result.Print(txtConcentrations, largefont, 17, 1, b, true, true, false, true)

	b.Min.Y += 36
	if haveAmbient {
		result.Print(fmt.Sprintf("%.2f%%  %.1fC\n%.3f kPa", rh, temperature, pressure/1000), titlefont, 9, 0, b, true, true, false, true)
	}

	return result
}

//func ViewPlot()
