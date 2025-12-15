package utils

import (
	"fmt"
	"time"
)

// ParseDate parses a date string in various formats
func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"02/01/2006",
		"01/02/2006",
		"2006-1-2",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseMonth parses a month string (e.g., "2024-01", "01/2024")
func ParseMonth(monthStr string) (time.Time, error) {
	formats := []string{
		"2006-01",
		"2006/01",
		"01/2006",
		"2006-1",
		"2006/1",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, monthStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse month: %s", monthStr)
}

// GetMonthStartEnd returns the start and end of a month
func GetMonthStartEnd(year int, month time.Month) (start, end time.Time) {
	start = time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end = time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.UTC)
	return start, end
}

// FormatDate formats a date for display
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatMonth formats a month for display
func FormatMonth(t time.Time) string {
	return t.Format("2006-01")
}

