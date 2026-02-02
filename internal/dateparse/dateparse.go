package dateparse

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDateTimeNow parses s using the current local time as the reference for relative expressions.
func ParseDateTimeNow(s string) (time.Time, error) {
	return ParseDateTime(s, time.Now())
}

// ParseDateTime parses RFC3339, YYYY-MM-DD, or relative expressions like yesterday, 2h ago, or monday.
func ParseDateTime(s string, now time.Time) (time.Time, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	normalized := strings.ToLower(raw)
	normalized = strings.TrimSpace(strings.Trim(normalized, ".,"))

	switch normalized {
	case "now":
		return now, nil
	case "today":
		return startOfDay(now), nil
	case "yesterday":
		return startOfDay(now.AddDate(0, 0, -1)), nil
	case "tomorrow":
		return startOfDay(now.AddDate(0, 0, 1)), nil
	}

	if d, ok := parseRelativeDuration(normalized); ok {
		return now.Add(-d), nil
	}

	if t, ok := parseWeekday(normalized, now); ok {
		return t, nil
	}

	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}

	if t, err := time.Parse("2006-01-02", raw); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date %q", raw)
}

func startOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func parseRelativeDuration(input string) (time.Duration, bool) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, false
	}

	if strings.HasSuffix(trimmed, "ago") {
		trimmed = strings.TrimSpace(strings.TrimSuffix(trimmed, "ago"))
	}
	if trimmed == "" {
		return 0, false
	}

	if d, err := time.ParseDuration(trimmed); err == nil {
		return d, true
	}

	// Support single-unit durations with day/week suffixes (e.g., 2d, 1w).
	num, unit := splitNumberUnit(trimmed)
	if num == "" || unit == "" {
		return 0, false
	}

	value, err := strconv.Atoi(num)
	if err != nil {
		return 0, false
	}

	switch unit {
	case "s":
		return time.Duration(value) * time.Second, true
	case "m":
		return time.Duration(value) * time.Minute, true
	case "h":
		return time.Duration(value) * time.Hour, true
	case "d":
		return time.Duration(value) * 24 * time.Hour, true
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func splitNumberUnit(s string) (string, string) {
	var i int
	for i < len(s) {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		i++
	}
	if i == 0 {
		return "", ""
	}
	num := s[:i]
	unit := strings.TrimSpace(s[i:])
	return num, unit
}

func parseWeekday(input string, now time.Time) (time.Time, bool) {
	weekday, ok := weekdayAliases[input]
	if !ok {
		return time.Time{}, false
	}

	daysAgo := (int(now.Weekday()) - int(weekday) + 7) % 7
	target := now.AddDate(0, 0, -daysAgo)
	return startOfDay(target), true
}

var weekdayAliases = map[string]time.Weekday{
	"sun":       time.Sunday,
	"sunday":    time.Sunday,
	"mon":       time.Monday,
	"monday":    time.Monday,
	"tue":       time.Tuesday,
	"tues":      time.Tuesday,
	"tuesday":   time.Tuesday,
	"wed":       time.Wednesday,
	"weds":      time.Wednesday,
	"wednesday": time.Wednesday,
	"thu":       time.Thursday,
	"thur":      time.Thursday,
	"thurs":     time.Thursday,
	"thursday":  time.Thursday,
	"fri":       time.Friday,
	"friday":    time.Friday,
	"sat":       time.Saturday,
	"saturday":  time.Saturday,
}
