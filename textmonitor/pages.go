package textmonitor

import (
	"fmt"
	"strings"
)

// PageMenuEntry is on page.. usually letters for activating some functionalities like enter IP number
type PageMenuEntry struct {
	Letter string
	Text   string
}

// Page is place to print and view printout. Scroll up and down
type Page struct {
	MenuCaption    string
	Content        string
	ScrollPosition int
	PageMenu       []PageMenuEntry
}

func (p *Page) ScrollToEnd(rowCount int) {
	p.ScrollPosition = max(0, len(strings.Split(p.Content, "\n"))-rowCount)
}

// Printout page part
func (p *Page) Printout(rowCount int, colCount int) string {
	rows := strings.Split(p.Content, "\n")
	fmt.Printf("Rows %#v\n", rows)
	p.ScrollPosition = max(0, p.ScrollPosition)
	p.ScrollPosition = min(len(rows)-1, p.ScrollPosition)

	rows = rows[p.ScrollPosition:]

	//cut rows
	var sb strings.Builder
	numberOfPrintedRows := 0
	emptyline := PadStringToLength(" ", colCount) + "\n"
	for i, row := range rows {
		if rowCount-2 <= i {
			break
		}
		if len(row) == 0 {
			sb.WriteString(emptyline)
			numberOfPrintedRows++
			continue
		}

		s := row[0:min(colCount, len(row)-1)]
		s = PadStringToLength(s, colCount)
		s = addNewLineIsNot(s)
		sb.WriteString(s)
		numberOfPrintedRows++
	}
	emptyLinesNeeded := rowCount - numberOfPrintedRows - 3
	if 0 < emptyLinesNeeded {
		sb.WriteString(strings.Repeat(emptyline, emptyLinesNeeded))
	}

	return sb.String()
}

type TitleStatus byte

const (
	Normal         TitleStatus = 0
	NoConnectivity TitleStatus = 1
	Warning        TitleStatus = 2
	Fail           TitleStatus = 3
)

func (p *TitleStatus) ColorEscape() string {
	code, haz := map[TitleStatus]string{
		Normal:         "\x1b[92;102;1m",
		NoConnectivity: "\x1b[35;3m ",
		Warning:        "\x1b[33;3m",
		Fail:           "\x1b[31;3m",
	}[*p]

	if !haz {
		return ""
	}
	return code
}

// Pages contains page
type Pages struct {
	Title      string
	Status     TitleStatus
	ActivePage int
	Items      []Page
}

func (p *Pages) Printout(rowCount int, colCount int) string {
	var sb strings.Builder

	sb.WriteString(ESCAPE_CLEAR)
	sb.WriteString(p.Status.ColorEscape())
	sb.WriteString(PadStringToLength(p.Title, colCount))
	sb.WriteString(ESCAPE_COLORBACKDEFAULT)

	//clamp
	p.ActivePage = min(p.ActivePage, len(p.Items)-1)
	p.ActivePage = max(0, p.ActivePage)

	sb.WriteString(p.Items[p.ActivePage].Printout(rowCount, colCount))

	menucolor := "\x1b[92;102;1m"
	//sb.WriteString()

	arr := make([]string, len(p.Items))
	for i, m := range p.Items {
		arr[i] = fmt.Sprintf("%v %s", i, m.MenuCaption)
	}
	arr = padArrItems(arr, colCount)
	for _, s := range arr {
		//first space is in between number and text
		sb.WriteString(ESCAPE_STATUSBAR_YELLOW + strings.Replace(s, " ", " "+menucolor, 1) + ESCAPE_COLORBACKDEFAULT)
	}

	return strings.ReplaceAll(sb.String(), "\n", "\r\n")
}
