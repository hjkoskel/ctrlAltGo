/*
Library for creating text based monitoring system

Prints status of system in virtual windows.. On terminal or serial port term
*/
package textmonitor

import (
	"strings"
)

const (
	ESCAPE_CLEAR      = "\x1b[2J\x1b[H"
	ESCAPE_NORMALTEXT = "\x1b[30;40m"
	ESCAPE_PROMPTTEXT = "\x1b[37m"
	ESCAPE_SIDEMENU   = "\x1b[33;40m"

	ESCAPE_PREDICTED        = "\x1b[32;40;1m"
	ESCAPE_TITLETEXT        = "\x1b[37;100;1m"
	ESCAPE_COLORBACKDEFAULT = "\x1b[0m"
)

const (
	ESCAPE_STATUSBAR_GREEN     = "\x1b[92;102;1m"
	ESCAPE_STATUSBAR_GREENDARK = "\x1b[32;42;1m"
	ESCAPE_STATUSBAR_YELLOW    = "\x1b[93;103;1m"
	ESCAPE_STATUSBAR_RED       = "\x1b[91;101;1m"
	ESCAPE_STATUSBAR_GRAY      = "\x1b[37;47;1m"
	ESCAPE_STATUSBAR_MENUROW   = "\x1b[36;46;1m"
)

func addNewLineIsNot(s string) string {
	if len(s) == 0 {
		return "\n"
	}
	if string(s[len(s)-1]) == "\n" {
		return s
	}
	return s + "\n"
}

func PadStringToLength(input string, desiredLength int) string {
	padding := desiredLength - len(input)
	if padding <= 0 {
		return input[0 : desiredLength-1]
	}
	return input + strings.Repeat(" ", padding)
}

func padArrItems(arr []string, totalWanted int) []string {
	if len(arr) == 0 {
		return arr
	}
	nPerItem := totalWanted / len(arr)
	nLeft := totalWanted % len(arr)

	result := make([]string, len(arr))
	for i, s := range arr {
		result[i] = PadStringToLength(s, nPerItem)
	}

	if nLeft == 0 || len(result) == 1 {
		return result
	}
	result[0] = result[0] + strings.Repeat(" ", nLeft/2)
	endPad := nLeft / 2
	if nLeft%2 != 0 { //odd,
		endPad += 1
	}

	result[len(result)-1] = result[len(result)-1] + strings.Repeat(" ", endPad)
	return result
}
