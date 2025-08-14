package manager

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses a duration string with support for various formats
// Supports: 30m, 2h, 1h30m, 90 (seconds), 1.5h, etc.
func ParseDuration(s string) (time.Duration, error) {
	// If it's just a number, treat it as seconds
	if num, err := strconv.Atoi(s); err == nil {
		return time.Duration(num) * time.Second, nil
	}

	// Try standard Go duration parsing first
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Handle formats like "1h30m" or "2h 30m"
	s = strings.ReplaceAll(s, " ", "")

	// Handle decimal hours (e.g., "1.5h")
	if strings.Contains(s, ".") && strings.HasSuffix(s, "h") {
		hoursStr := strings.TrimSuffix(s, "h")
		if hours, err := strconv.ParseFloat(hoursStr, 64); err == nil {
			return time.Duration(hours * float64(time.Hour)), nil
		}
	}

	// Try parsing with regex for complex formats
	re := regexp.MustCompile(`(\d+)([hms])`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	var total time.Duration
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %s", match[1])
		}

		switch match[2] {
		case "h":
			total += time.Duration(value) * time.Hour
		case "m":
			total += time.Duration(value) * time.Minute
		case "s":
			total += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("unknown time unit: %s", match[2])
		}
	}

	if total == 0 {
		return 0, fmt.Errorf("duration cannot be zero")
	}

	return total, nil
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "unlimited"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	parts := []string{}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, " ")
}
