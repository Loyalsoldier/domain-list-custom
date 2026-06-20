package lib

import (
	"strings"
)

// IsEmpty checks if a string is empty after trimming spaces
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// RemoveComment removes comments from a line
func RemoveComment(line string) string {
	idx := strings.Index(line, "#")
	if idx == -1 {
		return line
	}
	return strings.TrimSpace(line[:idx])
}
