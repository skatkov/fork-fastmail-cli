package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func confirmPrompt(w io.Writer, prompt string, accepted ...string) (bool, error) {
	fmt.Fprint(w, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("failed to read confirmation: %w", err)
		}
		return false, fmt.Errorf("cancelled")
	}
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))
	for _, ok := range accepted {
		if response == ok {
			return true, nil
		}
	}
	return false, nil
}
