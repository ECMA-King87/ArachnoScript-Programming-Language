package main

import (
	"path/filepath"
	"strings"
)

func RealPath(path string) string {
	full_path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	if !pathExists(full_path) {
		panic("the system could not find the path specified: " + path)
	}
	return full_path
}

func IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

func RelativePath(base, target string) string {
	if !IsAbs(base) {
		base, _ = filepath.Abs(base)
	}
	if !IsAbs(target) {
		target, _ = filepath.Abs(target)
	}
	path, err := filepath.Rel(base, target)
	if err != nil {
		panic(err)
	}
	return path
}

func AbsPath(path string) string {
	target, _ := filepath.Abs(path)
	return target
}

func RelativePathToFile(file, target string) string {
	file_path := AbsPath(file)
	index := strings.LastIndex(file_path, "\\")
	substr := ""
	for i := 0; i <= index; i++ {
		ch := file_path[i]
		substr += string(ch)
	}
	return substr + filepath.Clean(target)
}
