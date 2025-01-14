package inputdev

import (
	"fmt"
	"strings"
)

const (
	KEY_RESERVED         KeyCode = 0
	KEY_ESC              KeyCode = 1
	KEY_1                KeyCode = 2
	KEY_2                KeyCode = 3
	KEY_3                KeyCode = 4
	KEY_4                KeyCode = 5
	KEY_5                KeyCode = 6
	KEY_6                KeyCode = 7
	KEY_7                KeyCode = 8
	KEY_8                KeyCode = 9
	KEY_9                KeyCode = 10
	KEY_0                KeyCode = 11
	KEY_MINUS            KeyCode = 12
	KEY_EQUAL            KeyCode = 13
	KEY_BACKSPACE        KeyCode = 14
	KEY_TAB              KeyCode = 15
	KEY_Q                KeyCode = 16
	KEY_W                KeyCode = 17
	KEY_E                KeyCode = 18
	KEY_R                KeyCode = 19
	KEY_T                KeyCode = 20
	KEY_Y                KeyCode = 21
	KEY_U                KeyCode = 22
	KEY_I                KeyCode = 23
	KEY_O                KeyCode = 24
	KEY_P                KeyCode = 25
	KEY_LEFTBRACE        KeyCode = 26
	KEY_RIGHTBRACE       KeyCode = 27
	KEY_ENTER            KeyCode = 28
	KEY_LEFTCTRL         KeyCode = 29
	KEY_A                KeyCode = 30
	KEY_S                KeyCode = 31
	KEY_D                KeyCode = 32
	KEY_F                KeyCode = 33
	KEY_G                KeyCode = 34
	KEY_H                KeyCode = 35
	KEY_J                KeyCode = 36
	KEY_K                KeyCode = 37
	KEY_L                KeyCode = 38
	KEY_SEMICOLON        KeyCode = 39 //Ö
	KEY_APOSTROPHE       KeyCode = 40
	KEY_GRAVE            KeyCode = 41
	KEY_LEFTSHIFT        KeyCode = 42
	KEY_BACKSLASH        KeyCode = 43
	KEY_Z                KeyCode = 44
	KEY_X                KeyCode = 45
	KEY_C                KeyCode = 46
	KEY_V                KeyCode = 47
	KEY_B                KeyCode = 48
	KEY_N                KeyCode = 49
	KEY_M                KeyCode = 50
	KEY_COMMA            KeyCode = 51
	KEY_DOT              KeyCode = 52
	KEY_SLASH            KeyCode = 53
	KEY_RIGHTSHIFT       KeyCode = 54
	KEY_KPASTERISK       KeyCode = 55
	KEY_LEFTALT          KeyCode = 56
	KEY_SPACE            KeyCode = 57
	KEY_CAPSLOCK         KeyCode = 58
	KEY_F1               KeyCode = 59
	KEY_F2               KeyCode = 60
	KEY_F3               KeyCode = 61
	KEY_F4               KeyCode = 62
	KEY_F5               KeyCode = 63
	KEY_F6               KeyCode = 64
	KEY_F7               KeyCode = 65
	KEY_F8               KeyCode = 66
	KEY_F9               KeyCode = 67
	KEY_F10              KeyCode = 68
	KEY_NUMLOCK          KeyCode = 69
	KEY_SCROLLLOCK       KeyCode = 70
	KEY_KP7              KeyCode = 71
	KEY_KP8              KeyCode = 72
	KEY_KP9              KeyCode = 73
	KEY_KPMINUS          KeyCode = 74
	KEY_KP4              KeyCode = 75
	KEY_KP5              KeyCode = 76
	KEY_KP6              KeyCode = 77
	KEY_KPPLUS           KeyCode = 78
	KEY_KP1              KeyCode = 79
	KEY_KP2              KeyCode = 80
	KEY_KP3              KeyCode = 81
	KEY_KP0              KeyCode = 82
	KEY_KPDOT            KeyCode = 83
	KEY_ZENKAKUHANKAKU   KeyCode = 85
	KEY_102ND            KeyCode = 86
	KEY_F11              KeyCode = 87
	KEY_F12              KeyCode = 88
	KEY_RO               KeyCode = 89
	KEY_KATAKANA         KeyCode = 90
	KEY_HIRAGANA         KeyCode = 91
	KEY_HENKAN           KeyCode = 92
	KEY_KATAKANAHIRAGANA KeyCode = 93
	KEY_MUHENKAN         KeyCode = 94
	KEY_KPJPCOMMA        KeyCode = 95
	KEY_KPENTER          KeyCode = 96
	KEY_RIGHTCTRL        KeyCode = 97
	KEY_KPSLASH          KeyCode = 98
	KEY_SYSRQ            KeyCode = 99
	KEY_RIGHTALT         KeyCode = 100
	KEY_LINEFEED         KeyCode = 101
	KEY_HOME             KeyCode = 102
	KEY_UP               KeyCode = 103
	KEY_PAGEUP           KeyCode = 104
	KEY_LEFT             KeyCode = 105
	KEY_RIGHT            KeyCode = 106
	KEY_END              KeyCode = 107
	KEY_DOWN             KeyCode = 108
	KEY_PAGEDOWN         KeyCode = 109
	KEY_INSERT           KeyCode = 110
	KEY_DELETE           KeyCode = 111
	KEY_MACRO            KeyCode = 112
	KEY_MUTE             KeyCode = 113
	KEY_VOLUMEDOWN       KeyCode = 114
	KEY_VOLUMEUP         KeyCode = 115
	KEY_POWER            KeyCode = 116
	KEY_KPEQUAL          KeyCode = 117
	KEY_KPPLUSMINUS      KeyCode = 118
	KEY_PAUSE            KeyCode = 119
	KEY_SCALE            KeyCode = 120
	KEY_KPCOMMA          KeyCode = 121
	KEY_HANGEUL          KeyCode = 122
	KEY_HANGUEL          KeyCode = KEY_HANGEUL
	KEY_HANJA            KeyCode = 123
	KEY_YEN              KeyCode = 124
	KEY_LEFTMETA         KeyCode = 125
	KEY_RIGHTMETA        KeyCode = 126
	KEY_COMPOSE          KeyCode = 127
	KEY_STOP             KeyCode = 128
	KEY_AGAIN            KeyCode = 129
	KEY_PROPS            KeyCode = 130
	KEY_UNDO             KeyCode = 131
	KEY_FRONT            KeyCode = 132
	KEY_COPY             KeyCode = 133
	KEY_OPEN             KeyCode = 134
	KEY_PASTE            KeyCode = 135
	KEY_FIND             KeyCode = 136
	KEY_CUT              KeyCode = 137
	KEY_HELP             KeyCode = 138
	KEY_MENU             KeyCode = 139
	KEY_CALC             KeyCode = 140
	KEY_SETUP            KeyCode = 141
	KEY_SLEEP            KeyCode = 142
	KEY_WAKEUP           KeyCode = 143
	KEY_FILE             KeyCode = 144
	KEY_SENDFILE         KeyCode = 145
	KEY_DELETEFILE       KeyCode = 146
	KEY_XFER             KeyCode = 147
	KEY_PROG1            KeyCode = 148
	KEY_PROG2            KeyCode = 149
	KEY_WWW              KeyCode = 150
	KEY_MSDOS            KeyCode = 151
	KEY_COFFEE           KeyCode = 152
	KEY_SCREENLOCK       KeyCode = KEY_COFFEE
	KEY_ROTATE_DISPLAY   KeyCode = 153
	KEY_DIRECTION        KeyCode = KEY_ROTATE_DISPLAY
	KEY_CYCLEWINDOWS     KeyCode = 154
	KEY_MAIL             KeyCode = 155
	KEY_BOOKMARKS        KeyCode = 156
	KEY_COMPUTER         KeyCode = 157
	KEY_BACK             KeyCode = 158
	KEY_FORWARD          KeyCode = 159
	KEY_CLOSECD          KeyCode = 160
	KEY_EJECTCD          KeyCode = 161
	KEY_EJECTCLOSECD     KeyCode = 162
	KEY_NEXTSONG         KeyCode = 163
	KEY_PLAYPAUSE        KeyCode = 164
	KEY_PREVIOUSSONG     KeyCode = 165
	KEY_STOPCD           KeyCode = 166
	KEY_RECORD           KeyCode = 167
	KEY_REWIND           KeyCode = 168
	KEY_PHONE            KeyCode = 169
	KEY_ISO              KeyCode = 170
	KEY_CONFIG           KeyCode = 171
	KEY_HOMEPAGE         KeyCode = 172
	KEY_REFRESH          KeyCode = 173
	KEY_EXIT             KeyCode = 174
	KEY_MOVE             KeyCode = 175
	KEY_EDIT             KeyCode = 176
	KEY_SCROLLUP         KeyCode = 177
	KEY_SCROLLDOWN       KeyCode = 178
	KEY_KPLEFTPAREN      KeyCode = 179
	KEY_KPRIGHTPAREN     KeyCode = 180
	KEY_NEW              KeyCode = 181
	KEY_REDO             KeyCode = 182
	KEY_F13              KeyCode = 183
	KEY_F14              KeyCode = 184
	KEY_F15              KeyCode = 185
	KEY_F16              KeyCode = 186
	KEY_F17              KeyCode = 187
	KEY_F18              KeyCode = 188
	KEY_F19              KeyCode = 189
	KEY_F20              KeyCode = 190
	KEY_F21              KeyCode = 191
	KEY_F22              KeyCode = 192
	KEY_F23              KeyCode = 193
	KEY_F24              KeyCode = 194
	KEY_PLAYCD           KeyCode = 200
	KEY_PAUSECD          KeyCode = 201
	KEY_PROG3            KeyCode = 202
	KEY_PROG4            KeyCode = 203
	KEY_DASHBOARD        KeyCode = 204
	KEY_SUSPEND          KeyCode = 205
	KEY_CLOSE            KeyCode = 206
	KEY_PLAY             KeyCode = 207
	KEY_FASTFORWARD      KeyCode = 208
	KEY_BASSBOOST        KeyCode = 209
	KEY_PRINT            KeyCode = 210
	KEY_HP               KeyCode = 211
	KEY_CAMERA           KeyCode = 212
	KEY_SOUND            KeyCode = 213
	KEY_QUESTION         KeyCode = 214
	KEY_EMAIL            KeyCode = 215
	KEY_CHAT             KeyCode = 216
	KEY_SEARCH           KeyCode = 217
	KEY_CONNECT          KeyCode = 218
	KEY_FINANCE          KeyCode = 219
	KEY_SPORT            KeyCode = 220
	KEY_SHOP             KeyCode = 221
	KEY_ALTERASE         KeyCode = 222
	KEY_CANCEL           KeyCode = 223
	KEY_BRIGHTNESSDOWN   KeyCode = 224
	KEY_BRIGHTNESSUP     KeyCode = 225
	KEY_MEDIA            KeyCode = 226
	KEY_SWITCHVIDEOMODE  KeyCode = 227
	KEY_KBDILLUMTOGGLE   KeyCode = 228
	KEY_KBDILLUMDOWN     KeyCode = 229
	KEY_KBDILLUMUP       KeyCode = 230
	KEY_SEND             KeyCode = 231
	KEY_REPLY            KeyCode = 232
	KEY_FORWARDMAIL      KeyCode = 233
	KEY_SAVE             KeyCode = 234
	KEY_DOCUMENTS        KeyCode = 235
	KEY_BATTERY          KeyCode = 236
	KEY_BLUETOOTH        KeyCode = 237
	KEY_WLAN             KeyCode = 238
	KEY_UWB              KeyCode = 239
	KEY_UNKNOWN          KeyCode = 240
	KEY_VIDEO_NEXT       KeyCode = 241
	KEY_VIDEO_PREV       KeyCode = 242
	KEY_BRIGHTNESS_CYCLE KeyCode = 243
	KEY_BRIGHTNESS_AUTO  KeyCode = 244
	KEY_BRIGHTNESS_ZERO  KeyCode = KEY_BRIGHTNESS_AUTO
	KEY_DISPLAY_OFF      KeyCode = 245
	KEY_WWAN             KeyCode = 246
	KEY_WIMAX            KeyCode = KEY_WWAN
	KEY_RFKILL           KeyCode = 247
	KEY_MICMUTE          KeyCode = 248
)

type Keyboard struct { //Keep status. No hardware connection inside. Easier to unit test
	State  map[KeyCode]uint32 //0= not pressed
	Buffer []KeyCode          //What was pressed... Separate function to read out keyboard buffer in text etc...
	Text   string             //Built while pressing
	//TODO capslock? not supported in embedded world
	Keymap   map[KeyCode]string
	Keyshift map[KeyCode]string

	CapsOn bool
}

func (p *Keyboard) PressedKeys() []KeyCode {
	result := []KeyCode{}
	for code, value := range p.State {
		if 0 < value {
			result = append(result, code)
		}
	}
	return result
}

var loadedKeymapShift = map[KeyCode]string{
	KEY_1: "!",
	KEY_2: "\"",
	KEY_3: "#",
	KEY_4: "¤",
	KEY_5: "%",
	KEY_6: "&",
	KEY_7: "/",
	KEY_8: "(",
	KEY_9: ")",
	KEY_0: "=",
}

// TODO use embed or eany mean suitable
var loadedKeymap = map[KeyCode]string{
	KEY_1:          "1",
	KEY_2:          "2",
	KEY_3:          "3",
	KEY_4:          "4",
	KEY_5:          "5",
	KEY_6:          "6",
	KEY_7:          "7",
	KEY_8:          "8",
	KEY_9:          "9",
	KEY_0:          "0",
	KEY_MINUS:      "-",
	KEY_EQUAL:      "=",
	KEY_BACKSPACE:  "",
	KEY_TAB:        "\t",
	KEY_Q:          "Q",
	KEY_W:          "W",
	KEY_E:          "E",
	KEY_R:          "R",
	KEY_T:          "T",
	KEY_Y:          "Y",
	KEY_U:          "U",
	KEY_I:          "I",
	KEY_O:          "O",
	KEY_P:          "P",
	KEY_LEFTBRACE:  "[",
	KEY_RIGHTBRACE: "]",
	KEY_ENTER:      "\n",
	KEY_A:          "A",
	KEY_S:          "S",
	KEY_D:          "D",
	KEY_F:          "F",
	KEY_G:          "G",
	KEY_H:          "H",
	KEY_J:          "J",
	KEY_K:          "K",
	KEY_L:          "L",
	KEY_SEMICOLON:  ";",
	KEY_APOSTROPHE: "'",
	KEY_GRAVE:      "`",
	KEY_BACKSLASH:  "\\",
	KEY_Z:          "Z",
	KEY_X:          "X",
	KEY_C:          "C",
	KEY_V:          "V",
	KEY_B:          "B",
	KEY_N:          "N",
	KEY_M:          "M",
	KEY_COMMA:      ",",
	KEY_DOT:        ".",
	KEY_SLASH:      "/",
	KEY_SPACE:      " ",
}

func InitKeyboard() Keyboard {
	return Keyboard{
		State:    make(map[KeyCode]uint32),
		Buffer:   []KeyCode{},
		Keymap:   loadedKeymap, //For printable buttons
		Keyshift: loadedKeymapShift,
	}
}

func (p *Keyboard) Shift() bool {
	key, _ := p.State[KEY_LEFTSHIFT]
	if 0 < key {
		return true
	}
	key, _ = p.State[KEY_RIGHTSHIFT]
	if 0 < key {
		return true
	}
	return false
}

func (p *Keyboard) UpperMode() bool { //TODO check caps?
	return p.CapsOn || p.Shift()
}

func (p *Keyboard) Update(ev *RawInputEvent) {
	if ev.Type != EV_KEY {
		return //Not reacting
	}
	fmt.Printf("\n\n---%v=%v----\n", ev.Code, ev.Value)
	p.State[KeyCode(ev.Code)] = ev.Value

	if ev.Value == 0 {
		return //Button released
	}
	if KeyCode(ev.Code) == KEY_CAPSLOCK {
		p.CapsOn = !p.CapsOn
	}
	p.Buffer = append(p.Buffer, KeyCode(ev.Code))

	s, hazPrint := p.Keymap[KeyCode(ev.Code)]
	if p.Shift() {
		fmt.Printf("HAVE SHIFT\n")
		sshift, hazShifted := p.Keyshift[KeyCode(ev.Code)]
		if hazShifted {
			s = sshift
			hazPrint = true
		}
	}

	if hazPrint {
		if p.UpperMode() {
			p.Text += strings.ToUpper(s)
		} else {
			p.Text += strings.ToLower(s)
		}
	}

}
