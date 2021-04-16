package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readSimpleField(line string) (string, string, error) {
	parts := strings.SplitN(line, ",", 2)
	switch len(parts) {
	case 2:
		return parts[0], parts[1], nil
	case 1:
		return parts[0], "", io.EOF
	default:
		return line, "", io.EOF
	}
}

func readComplexField(line string) (string, string, error) {
	stack := make([]byte, 1)
	stack[0] = line[0]

	for i := 1; i < len(line); i++ {
		c := line[i]
		top := stack[len(stack)-1]

		if c == '{' || c == '[' {
			stack = append(stack, c)
		} else if (c == '}' && top == '{') || (c == ']' && top == '[') {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			next := i + 1
			if next == len(line) {
				return line[0:next], "", nil
			} else if line[next] == ',' {
				return line[0:next], line[i+2:], nil
			} else {
				return "", "", errors.New("unbalanced parentheses")
			}
		}
	}

	return "", "", errors.New("unbalanced parentheses")
}

func readField(line string) (string, string, error) {
	if len(line) == 0 {
		return "", "", io.EOF
	}
	if line[0] == '{' || line[0] == '[' {
		return readComplexField(line)
	} else {
		return readSimpleField(line)
	}
}

func processLine(line string, writer *csv.Writer, line_no int) {
	cells := make([]string, 0)
	for {
		var cell string
		var err error
		cell, line, err = readField(line)
		cells = append(cells, cell)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Encountered error '%s' when processing line %d, skipping.\n", err, line_no)
			return
		}
	}
	writer.Write(cells)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	writer := csv.NewWriter(os.Stdout)

	line_no := 0
	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		processLine(strings.TrimSpace(text), writer, line_no)
		line_no += 1
	}
	writer.Flush()
}
