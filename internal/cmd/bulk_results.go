package cmd

import "fmt"

func printBulkResults(action, target string, succeededCount, failedCount int, failed map[string]string) {
	if failedCount == 0 {
		if target != "" {
			fmt.Printf("%s %d %s\n", action, succeededCount, target)
			return
		}
		fmt.Printf("%s %d\n", action, succeededCount)
		return
	}

	if target != "" {
		fmt.Printf("%s %d %s, %d failed:\n", action, succeededCount, target, failedCount)
	} else {
		fmt.Printf("%s %d, %d failed:\n", action, succeededCount, failedCount)
	}
	for id, errMsg := range failed {
		fmt.Printf("  %s: %s\n", id, errMsg)
	}
}
