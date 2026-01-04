package main

import (
	"fmt"
	"strings"
)

func SyntaxError(message string) string {
	return "\x1b[31mSyntaxError\x1b[0m: " + message
}

func SourceAtPosition(path string, line int, char int) string {
	return fmt.Sprintf("\r\nat (\x1b[34m%s\x1b[0m\x1b[33m:%d:%d\x1b[0m)", path, line, char)
}

// path: is either path to file or a source
//
// line: line in source to start from
//
// pos:  position of character to put "^" under
//
// count: the length of the token or the number of times to repeat "^"
//
// chars: limit the number of characters to be displayed on a line
//
// _range: number of lines from [line] to display
func SourceWithinRange(
	path string,
	line int,
	pos int,
	count int,
	// chars int,
	// _range int,
	source string,
) string {
	var lines []string
	if len(source) == 0 && pathExists(path) {
		lines = strings.Split(ReadTextFile(path), "\r\n")
	} else {
		lines = []string{source}
	}
	if line < 0 ||
		pos < 0 {
		// ||
		// _range < 0 ||
		// chars < 0
		throwMessage(
			"numeric arguments of sourceWithinRange must be greater than 0",
		)
	}
	sourceAtRange := ""
	for index := range lines {
		if len(sourceAtRange) > 0 {
			// the range has already been taken
			break
		}
		// if _range > 0 && len(sourceAtRange) > 0 {
		// 	// the range has already been taken
		// 	break
		// }
		if (index + 1) == line {
			// if chars > 0 {
			// 	// sourceAtRange = source[index].slice(0, chars)
			// 	sourceAtRange, _, _ = strings.Cut(lines[index], string(lines[index][chars]))
			// 	break
			// }
			array := make([]int, pos-1)
			line_source := ""
			for range array {
				line_source += " "
			}
			line_source += "\x1b[31m" + strings.Repeat("^", count) + "\x1b[0m"
			sourceAtRange = fmt.Sprintf("\r\n%s\r\n%s", lines[index], line_source)
			// if _range == 0 {
			// 	break
			// }
		}
		if (index + 1) > line {
			sourceAtRange += "\r\n" + lines[index]
		}
	}
	return "\r\n" + sourceAtRange
}
