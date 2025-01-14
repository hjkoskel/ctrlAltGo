/*
common utilities... TODO separate sub repo? TODO or move initializing here
*/
package ctrlaltgo

import (
	"fmt"
	"os"
	"time"
)

// JamIfErr Prints error and jams so kernel panic is not caused. USE on development only. Find better way to report error and recover from it
func JamIfErr(err error) {
	if err == nil {
		return
	}
	fmt.Printf("\n%s\n\n", err)
	for err != nil {
		if os.Getpid() != 1 {
			os.Exit(-1)
		}
		time.Sleep(time.Second * 5)
	}
}
