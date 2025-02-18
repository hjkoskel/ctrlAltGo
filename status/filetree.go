/*
Filetree, getting selected JSON data from filesystem tree in JSON format for GUI

-No recursion GUI asks absolute dir
-No permissions etc..
-File type detection is nice
-Separeted with

Serve directory listings as template!
*/
package status

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type FilePipeEntry struct {
	Path string
}

func (f FilePipeEntry) Name() string { //for gui
	return path.Base(f.Path)
}

type FileSocketEntry struct {
	Name string //Name for GUI
	Path string
}

type FileLinkEntry struct {
	Path   string //Whole path of this file entry
	LinkTo string
}

func (f FileLinkEntry) Name() string { //for gui
	return path.Base(f.Path)
}

type FileDeviceEntryResolved struct {
	//Resolved values, looked up or read separately
	Username    string
	Groupname   string
	Description string //What kind of device
	MimeType    string
}
type FileDeviceEntry struct {
	Path    string
	Mode    os.FileMode
	Uid     uint32
	Gid     uint32
	ModTime time.Time
	//Looking up by these
	Major    int64
	Minor    int64
	Resolved FileDeviceEntryResolved
}

func (f FileDeviceEntry) ModTimeFormatted() string {
	//Intelligent way
	return FormatTimestampOnFile(f.ModTime, time.Now())
}

func (f FileDeviceEntry) Name() string { //for gui
	return path.Base(f.Path)
}

type FileEntryResolved struct {
	Username    string
	Groupname   string
	MimeType    string
	Previewdata []byte
}

type FileEntry struct {
	Path     string //Whole path to this file entry
	Mode     os.FileMode
	Uid      uint32
	Gid      uint32
	ModTime  time.Time
	Size     int64
	Resolved FileEntryResolved
}

func (f FileEntry) PathLink() string {
	return path.Join("/browse/", f.Path)
}

func (f FileEntry) PreviewLink() string {
	return path.Join("/browse/", path.Dir(f.Path)) + "?preview=" + path.Base(f.Path)
}

func (f FileEntry) Name() string {
	return path.Base(f.Path)
}

func FormatTimestampOnFile(tstamp time.Time, now time.Time) string {
	if tstamp.After(time.Now()) {
		return fmt.Sprintf("future %s", tstamp.Sub(now))
	}
	//if tstamp.Year() != now.Year() { //better to give full time and date
	return fmt.Sprintf("%02d:%02d:%02d: %d.%d.%d", tstamp.Hour(), tstamp.Minute(), tstamp.Second(), tstamp.Day(), tstamp.Month(), tstamp.Year())
	//}
}

func (f FileEntry) ModTimeFormatted() string {
	//Intelligent way
	return FormatTimestampOnFile(f.ModTime, time.Now())
}

type DirectoryEntry struct {
	Path        string //Absolute path of this dir
	Size        int64  //Total size what under this or this is leaf. If not traversed then this is zero
	Dirs        []DirectoryEntry
	DeviceFiles []FileDeviceEntry
	Files       []FileEntry
	Links       []FileLinkEntry
	NamedPipes  []FilePipeEntry
	Sockets     []FileSocketEntry

	Errors []error
}

func (f DirectoryEntry) PathLink() string {
	return path.Join("/browse/", f.Path)
}

func (f DirectoryEntry) Name() string {
	return path.Base(f.Path)
}

func ReadDirectoryEntry(rootfilename string, recursive bool) (DirectoryEntry, []error) {
	de, errRead := os.ReadDir(rootfilename)
	if errRead != nil {
		return DirectoryEntry{}, []error{errRead}
	}

	result := DirectoryEntry{
		Path:        rootfilename,
		Size:        0,
		Dirs:        []DirectoryEntry{},
		DeviceFiles: []FileDeviceEntry{},
		Files:       []FileEntry{},
		Links:       []FileLinkEntry{},
	}

	errList := []error{} //Can contain multiple fails, missing access rights etc..
	for _, d := range de {
		absName := path.Join(rootfilename, d.Name())
		if d.IsDir() {
			if !recursive {
				result.Dirs = append(result.Dirs, DirectoryEntry{Path: absName})
				continue
			}

			parsedDir, errParseDir := ReadDirectoryEntry(absName, recursive)
			if errParseDir != nil {
				errList = append(errList, fmt.Errorf("error reading dir %s err:%s", absName, errParseDir))
				continue
			}
			result.Dirs = append(result.Dirs, parsedDir)
			continue
		}
		stat, errStat := os.Stat(absName)
		if errStat != nil {
			errList = append(errList, fmt.Errorf("error getting stats on file %s err:%w", absName, errStat))
			continue
		}
		mode := stat.Mode()

		if mode.IsRegular() {
			//TODO NORMAL
			result.Files = append(result.Files, FileEntry{
				Path:    absName,
				Mode:    stat.Mode(),
				Uid:     0,
				Gid:     0,
				ModTime: stat.ModTime(),
				Size:    stat.Size()})
			continue
		}

		//--- Device file ----
		if mode&os.ModeDevice != 0 {
			result.DeviceFiles = append(result.DeviceFiles, FileDeviceEntry{
				Path:    absName,
				Mode:    stat.Mode(),
				Uid:     0,
				Gid:     0,
				ModTime: stat.ModTime(),
			})
			continue
		}
		//---  symlink ---
		if mode&os.ModeSymlink != 0 {
			actualPath, errLink := os.Readlink(absName)
			if errLink != nil {
				errList = append(errList, fmt.Errorf("error getting link for %s err:%s", absName, errLink))
				continue
			}
			result.Links = append(result.Links, FileLinkEntry{
				Path:   absName,
				LinkTo: actualPath,
			})
			continue
		}
		//--- Socket ---
		if mode&os.ModeSocket != 0 {
			result.Sockets = append(result.Sockets, FileSocketEntry{Path: absName})
			continue
		}
		//---Named pipes
		if mode&os.ModeNamedPipe != 0 {
			result.NamedPipes = append(result.NamedPipes, FilePipeEntry{Path: absName})
			continue
		}
	}
	result.Errors = errList
	return result, errList
}

func listDirs(dirname string) ([]string, error) {
	dirlisting, errDir := os.ReadDir(dirname)
	if errDir != nil {
		return nil, errDir
	}
	arr := []string{}
	for _, d := range dirlisting {
		if d.IsDir() {
			arr = append(arr, d.Name())
		}
	}
	return arr, nil
}

type TreeOpening struct {
	DirectoryPath string        `json:"directoryPath,omitempty"`
	Subs          []TreeOpening `json:"subs,omitempty"`
}

func (p *TreeOpening) ToUlList() string {
	var sb strings.Builder
	sb.WriteString("<li><a href=\"/browse" + p.DirectoryPath + "\">" + path.Base(p.DirectoryPath) + "</a>")

	if 0 < len(p.Subs) {
		sb.WriteString("\n<ul>")
		for _, sub := range p.Subs {
			sb.WriteString(sub.ToUlList())
		}
		sb.WriteString("\n</ul>")
	}
	sb.WriteString("</li>\n")
	return sb.String()
}

func ReadOpening(dirname string, continueTo []string) TreeOpening {
	lst, _ := listDirs(dirname)
	result := TreeOpening{DirectoryPath: dirname, Subs: make([]TreeOpening, len(lst))}
	for i, dn := range lst {
		result.Subs[i].DirectoryPath = path.Join(dirname, dn)
		if 0 < len(continueTo) {
			if continueTo[0] == dn {
				result.Subs[i] = ReadOpening(result.Subs[i].DirectoryPath, continueTo[1:])
			}
		}
	}

	return result
}

func ReadOpeningFromDir(startDir string) TreeOpening {
	parts := strings.Split(path.Clean(startDir), "/")

	return ReadOpening("/", parts)
}

/*
func ReadTreeOpening(startDir string) TreeOpening {
	parts := strings.Split(path.Clean(startDir), "/")
	fmt.Printf("PARTS %#v\n", parts)
	result := TreeOpening{}
	pathnow := "/"

	for _, s := range parts {
		result.DirectoryPath = pathnow
		lst, _ := listDirs(pathnow)
		result = append(result, lst)
		pathnow += s + "/"
	}
	if parts[0] == "." {
		return result
	}
	lst, _ := listDirs(pathnow)
	result = append(result, lst)
	return result
}*/

/*
// Show all siblings,childrens,  and root dir items
func ReadTreeOpening(startDir string) [][]string {
	parts := strings.Split(path.Clean(startDir), "/")
	fmt.Printf("PARTS %#v\n", parts)
	result := [][]string{}
	pathnow := "/"
	for _, s := range parts {
		fmt.Printf("Pathnow %s\n", s)
		lst, _ := listDirs(pathnow)
		result = append(result, lst)
		pathnow += s + "/"
	}
	if parts[0] == "." {
		return result
	}
	lst, _ := listDirs(pathnow)
	result = append(result, lst)
	return result
}
*/
