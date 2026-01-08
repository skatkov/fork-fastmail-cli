package cmd

import "fmt"

func printAlready(action string) {
	fmt.Println(action)
}

func formatAlready(template string, args ...any) string {
	return fmt.Sprintf(template, args...)
}
