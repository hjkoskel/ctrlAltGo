/*
Mouse keeps latest status, based on events recieved from device
*/
package inputdev

import (
	"image"
	"time"
)

type MouseClick struct {
	XY           image.Point
	LeftButton   bool
	MiddleButton bool
	RightButton  bool
	When         time.Time
}

type Mouse struct {
	LeftButton   bool
	MiddleButton bool
	RightButton  bool

	RelRange   image.Rectangle //clamp by this
	StartCoord image.Point     //touch start here
	RelCoord   image.Point

	Cursor     image.Point
	cursorWork image.Point

	Clicks []MouseClick

	Touching     bool
	touchChanged bool
}

func ClampWithRect(pt image.Point, rect image.Rectangle) image.Point {
	return image.Point{
		X: min(max(pt.X, rect.Min.X), rect.Max.X),
		Y: min(max(pt.Y, rect.Min.Y), rect.Max.Y),
	}
}

const (
	ABS_X KeyCode = 0x00
	ABS_Y KeyCode = 0x01

	ABS_MT_POSITION_X KeyCode = 0x35
	ABS_MT_POSITION_Y KeyCode = 0x36

	BTN_LEFT   KeyCode = 0x110
	BTN_RIGHT  KeyCode = 0x111
	BTN_MIDDLE KeyCode = 0x112

	BTN_TOUCH       KeyCode = 0x14a
	BTN_TOOL_FINGER KeyCode = 0x145
	//HUOM! syn paketti
)

func (p *Mouse) Update(ev *RawInputEvent) {
	switch ev.Type {
	case EV_ABS:
		switch ev.Code {
		/*case uint16(ABS_MT_POSITION_X):
			p.cursorWork.X = int(ev.Value)
		case uint16(ABS_MT_POSITION_Y):
			p.cursorWork.Y = int(ev.Value)*/
		case uint16(ABS_X):
			p.cursorWork.X = int(ev.Value)
		case uint16(ABS_Y):
			p.cursorWork.Y = int(ev.Value)
		}
	case EV_KEY:
		switch ev.Code {
		case uint16(BTN_LEFT):
			p.LeftButton = 0 < ev.Value
		case uint16(BTN_RIGHT):
			p.RightButton = 0 < ev.Value
		case uint16(BTN_MIDDLE):
			p.MiddleButton = 0 < ev.Value
		case uint16(BTN_TOUCH):
			touchNow := 0 < ev.Value
			p.touchChanged = !(touchNow == p.Touching)
			p.Touching = touchNow
		}
	case EV_SYN:
		p.Cursor = p.cursorWork
		click := MouseClick{
			XY:           image.Point{X: p.Cursor.X, Y: p.Cursor.Y},
			LeftButton:   p.LeftButton,
			MiddleButton: p.MiddleButton,
			RightButton:  p.RightButton,
			When:         ev.Timestamp}
		if len(p.Clicks) == 0 {
			if p.LeftButton || p.RightButton || p.MiddleButton { //Any key is pressed
				p.Clicks = append(p.Clicks, click)
			}
		} else {
			latest := p.Clicks[len(p.Clicks)-1]
			if latest.LeftButton != p.LeftButton || latest.RightButton != p.RightButton || latest.MiddleButton != p.MiddleButton {
				p.Clicks = append(p.Clicks, click)
			}
		}

		if p.touchChanged {
			//touch changed
			if !p.Touching {
				p.StartCoord = p.Cursor
			} else {
				p.RelCoord.X += p.Cursor.X - p.StartCoord.X
				p.RelCoord.Y += p.Cursor.Y - p.StartCoord.Y
			}
			p.RelCoord = ClampWithRect(p.RelCoord, p.RelRange)
		}
	}

}
