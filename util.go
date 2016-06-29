package gotools

import (
	"os"
	"path/filepath"
)

func assert(expression bool) {
	if !expression {
		panic("assert failed")
	}
}

func GetCurrentAbsPath() string {
	abs, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	return abs
}
