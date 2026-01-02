package desktop

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var desktopPlaceholders = map[string]bool{
	"%f": true,
	"%F": true,
	"%u": true,
	"%U": true,
	"%d": true,
	"%D": true,
	"%n": true,
	"%N": true,
	"%i": true,
	"%c": true,
	"%k": true,
	"%v": true,
	"%m": true,
}

func Launch(command string) {
	cmd := exec.Command("sh", "-c", command)
	log.Printf("Launching command: %s\n", command)

	if err := cmd.Start(); err != nil {
		log.Printf("Unable to launch command: %s, error: %v\n", command, err)
	}
}

func CleanExec(execLine string) (string, error) {
	args, err := splitCommandLine(execLine)
	if err != nil {
		return "", fmt.Errorf("failed to parse command line: %w", err)
	}

	var filteredArgs []string
	for _, arg := range args {
		if !containsPlaceholder(arg) {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	return strings.Join(filteredArgs, " "), nil
}

func containsPlaceholder(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '%' {
			placeholder := s[i : i+2]
			if desktopPlaceholders[placeholder] {
				return true
			}
		}
	}
	return false
}

func splitCommandLine(line string) ([]string, error) {
	var args []string
	var currentArg strings.Builder
	inQuotes := false
	escapeNext := false

	for i, r := range line {
		if escapeNext {
			currentArg.WriteRune(r)
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			if i < len(line)-1 {
				escapeNext = true
			}
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t', '\n', '\r':
			if inQuotes {
				currentArg.WriteRune(r)
			} else {
				if currentArg.Len() > 0 {
					args = append(args, currentArg.String())
					currentArg.Reset()
				}
			}
		default:
			currentArg.WriteRune(r)
		}
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	if inQuotes {
		return nil, fmt.Errorf("unclosed quotes in command line")
	}

	return args, nil
}
