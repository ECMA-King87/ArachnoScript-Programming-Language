package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func GetUserInput() string {
	input_reader := bufio.NewReader(os.Stdin)
	input, err := input_reader.ReadString('\n')
	if err != nil {
		throwError(err)
	}
	return strings.TrimSpace(input)
}

func Prompt(message string, _default string) (string, error) {
	fmt.Print(message)
	input_reader := bufio.NewReader(os.Stdin)
	input, err := input_reader.ReadString('\n')
	return strings.TrimSpace(input), err
}

func print(data ...any) {
	fmt.Print(data...)
}

func println(data ...any) {
	data = append(data, "\r\n")
	print(data...)
}

func printf(f string, data ...any) {
	fmt.Printf(f, data...)
}

func sprintf(s string, values ...any) string {
	return fmt.Sprintf(s, values...)
}

func sprint(values ...any) string {
	return fmt.Sprint(values...)
}

func throwError(err error) {
	log.Fatal(err)
}

func throwMessage(message string) {
	println(message)
	os.Exit(1)
}
