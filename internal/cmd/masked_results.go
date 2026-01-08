package cmd

import "fmt"

func printMaskedBulkResults(action string, succeeded, failed int, domain string, errors []string) {
	if failed == 0 {
		fmt.Printf("Successfully %s %d aliases for %s\n", action, succeeded, domain)
		return
	}

	fmt.Printf("Partially %s %d aliases, %d failed:\n", action, succeeded, failed)
	for _, err := range errors {
		fmt.Printf("  %s\n", err)
	}
}
