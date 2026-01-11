package main

import "os"

func pathExists(path string) bool {
	f, err := os.Open(path)
	f.Close()
	return err == nil
}

func ReadTextFile(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		throwError(err)
	}
	return string(bytes)
}

func CreateTempFile(pattern string) *os.File {
	file, err := os.CreateTemp("temp", pattern)
	if err != nil {
		throwError(err)
	}
	return file
}
