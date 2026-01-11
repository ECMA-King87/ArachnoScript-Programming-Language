package main

import (
	"os"
)

var AS = []string{
	`          _____                    _____          `,
	`         /\    \                  /\    \         `,
	`        /::\    \                /::\    \        `,
	`       /::::\    \              /::::\    \       `,
	`      /::::::\    \            /::::::\    \      `,
	`     /:::/\:::\    \          /:::/\:::\    \     `,
	`    /:::/__\:::\    \        /:::/__\:::\    \    `,
	`   /::::\   \:::\    \       \:::\   \:::\    \   `,
	`  /::::::\   \:::\    \    ___\:::\   \:::\    \  `,
	` /:::/\:::\   \:::\    \  /\   \:::\   \:::\    \ `,
	`/:::/  \:::\   \:::\____\/::\   \:::\   \:::\____\`,
	`\::/    \:::\  /:::/    /\:::\   \:::\   \::/    /`,
	` \/____/ \:::\/:::/    /  \:::\   \:::\   \/____/ `,
	`          \::::::/    /    \:::\   \:::\    \     `,
	`           \::::/    /      \:::\   \:::\____\    `,
	`           /:::/    /        \:::\  /:::/    /    `,
	`          /:::/    /          \:::\/:::/    /     `,
	`         /:::/    /            \::::::/    /      `,
	`        /:::/    /              \::::/    /       `,
	`        \::/    /                \::/    /        `,
	`         \/____/                  \/____/         `,
	`                                                  `,
}

var tempDirCreated = false

func initialize() {
	if !pathExists("temp") {
		println("creating temp directory...")
		err := os.Mkdir("temp", os.ModeDir)
		if err != nil {
			println("\x1b[35mError\x1b[0m: error creating temp directory")
		} else {
			println("directory created: temp")
		}
	} else {
		tempDirCreated = true
	}
}

func REPL() {
	var file *os.File
	var err error
	if !tempDirCreated {
		file, err = os.CreateTemp("", "repl-")
		if err != nil {
			throwError(err)
		}
	} else {
		file = CreateTempFile("repl-")
	}
	defer file.Close()
	display := [...]string{
		"ArachnoScript REPL - \x1b[32mv0.1\x1b[0m",
		"ARE v0.1",
		"enter .peace to exit the repl.",
	}
	for _, line := range AS {
		println(line)
	}
	for _, line := range display {
		println(line)
	}
	runtime := NewRuntime()
	stdEnv.sourcePath = AbsPath(file.Name())
	for {
		print("\x1b[32m>>\x1b[0m ")
		input := GetUserInput()
		file.WriteString(input + "\r\n")
		if input == ".peace" {
			break
		}
		var Parser *Parser = NewParser(file.Name(), "program", input)
		program := Parser.Parse(true)
		runtime.Evaluate(program, stdEnv)
	}
}

func RunScript(path string) {
	if !IsAbs(path) {
		path = AbsPath(path)
	}
	if !pathExists(path) {
		throwMessage("path: \x1b[31m" + path + "\x1b[0m; does not exist")
	}
	parser := NewParser(path, "program", "")
	program := parser.Parse(true)
	runtime := NewRuntime()
	env := NewEnv(stdEnv, "program", path)
	runtime.Evaluate(program, env)
}

var arguments = os.Args[1:]

var exec_path = RealPath(os.Args[0])

func main() {
	initialize()
	RunSTD("../stdlib/main.as")
}

var stdEnv *Environment

func RunSTD(path string) {
	path = RelativePathToFile(exec_path, path)
	parser := NewParser(path, "program", "")
	program := parser.Parse(true)
	runtime := NewRuntime()
	stdEnv = CreateScriptEnv(runtime, path)
	runtime.Evaluate(program, stdEnv)
}
