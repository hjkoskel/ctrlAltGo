/*
Clock app is just simple binary started by PID1 program

Just like on normal pc instead of having all set up code on this app.
PID1 program re-starts if this app fails
*/
package main

import (
	"fmt"
	"image"
	"time"

	"github.com/hjkoskel/gomonochromebitmap"
)

func createClockFace() gomonochromebitmap.MonoBitmap {
	result := gomonochromebitmap.NewMonoBitmap(128, 64, false)
	testfont := gomonochromebitmap.GetFont_8x8()
	result.Circle(image.Point{X: 64, Y: 32}, 32, true)

	tNow := time.Now()
	result.Print(fmt.Sprintf("%02v:%02v:%02v", tNow.Hour(), tNow.Minute(), tNow.Second()), testfont, 0, 1, result.Bounds(), true, true, false, false)
	return result
}

func main() {
	picrender := gomonochromebitmap.BlockGraphics{
		Clear:      false,
		HaveBorder: false,
		//BorderColor: gomonochromebitmap.FGANSI_YELLOW,
		//TextColor:   gomonochromebitmap.FGANSI_BLUE,
	}

	for {
		pic := createClockFace()
		fmt.Printf("\n\n\n%s\n", picrender.ToHalfBlockChars(&pic))
		time.Sleep(time.Second)
	}

}
