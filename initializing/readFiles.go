/*
Utility functions for reading /proc/sys files
Todo change to other dir depending on need
*/
package initializing

import (
	"os"
	"strconv"
	"strings"
)

// ReadIntFile
func ReadIntfile(filename string) (int64, error) {
	buf, errFile := os.ReadFile(filename)
	if errFile != nil {
		return 0, errFile
	}
	return strconv.ParseInt(strings.TrimSpace(string(buf)), 10, 64)
}
