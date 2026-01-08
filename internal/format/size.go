package format

import (
	"fmt"
	"math"
)

func FormatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// FormatBytes converts bytes to human-readable format (KB, MB, GB, TB).
func FormatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	exp := int(math.Log(float64(bytes)) / math.Log(unit))
	if exp > 4 {
		exp = 4
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(bytes) / math.Pow(unit, float64(exp))

	return fmt.Sprintf("%.1f %s", value, units[exp])
}
