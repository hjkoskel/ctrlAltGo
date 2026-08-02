/*
config.txt manipulation functionality
*/
package initializing

import (
	"fmt"
	"os"
	"strings"
)

func CommentInitramfsInConfigFile(fname string) error {
	fmt.Printf("initramfs file is located now at %s\n", fname)
	byt, errRead := os.ReadFile(fname)
	if errRead != nil {
		return fmt.Errorf("reading initramfs config.txt from %s failed err:%s", fname, errRead)
	}

	cleared := CommentInitramfsInConfigtxt(byt)

	fmt.Printf("---- UPDATED config.txt ----\n%s\n", cleared)

	errWrite := os.WriteFile(fname, cleared, 0666)
	if errWrite != nil {
		return fmt.Errorf("writing removed initramfs config.txt to %s failed err:%s", fname, errWrite)
	}
	return nil
}

func CommentInitramfsInConfigtxt(content []byte) []byte {
	rows := strings.Split(string(content), "\n")
	var sb strings.Builder
	for _, row := range rows {
		s := strings.TrimSpace(row)
		if strings.HasPrefix(s, "initramfs") {
			sb.WriteString("# ")
		}
		sb.WriteString(s)
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}
