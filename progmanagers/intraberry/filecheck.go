package main

import (
	"errors"
	"os"
)

func FileExists(fname string) bool {
	file, err := os.Open(fname)
	defer file.Close()
	return !errors.Is(err, os.ErrNotExist)
}
