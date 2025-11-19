package files

import (
	"crypto/md5"
	"strconv"
	"strings"

	"github.com/mkozhukh/tesei"
)

var base62Chars = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func hashBase62(input string, size int) string {
	if size <= 0 {
		size = 8
	} else if size > 11 {
		size = 11
	}

	sum := md5.Sum([]byte(input))

	// Take the first 8 bytes (64 bits)
	var val uint64
	for i := 0; i < 8; i++ {
		val = (val << 8) | uint64(sum[i])
	}

	chars := make([]byte, size)
	for i := size - 1; i >= 0; i-- {
		chars[i] = base62Chars[val%62]
		val /= 62
	}

	return string(chars)
}

// ResolveString replaces template variables in the format {{key}} with values from metadata
func ResolveString(input string, msg *tesei.Message[TextFile]) string {
	// Quick check - if no template markers, return immediately
	if !strings.Contains(input, "{{") {
		return input
	}

	var result strings.Builder
	result.Grow(len(input)) // Pre-allocate to avoid reallocations

	i := 0
	for i < len(input) {
		// Find next template marker
		start := strings.Index(input[i:], "{{")
		if start == -1 {
			// No more templates, append rest of string
			result.WriteString(input[i:])
			break
		}

		// Append text before template
		result.WriteString(input[i : i+start])
		i += start

		// Find closing marker
		end := strings.Index(input[i+2:], "}}")
		if end == -1 {
			// Malformed template, append rest as-is
			result.WriteString(input[i:])
			break
		}

		// Extract and resolve key
		key := input[i+2 : i+2+end]
		if key != "" {
			if value := FormatValue(msg.Metadata[key]); value != "" {
				result.WriteString(value)
			}
			// If value is empty or key doesn't exist, we write nothing (key disappears)
		}

		i += 2 + end + 2 // Move past "}}"
	}

	return result.String()
}

// FormatValue converts metadata values to strings with type safety
func FormatValue(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}
