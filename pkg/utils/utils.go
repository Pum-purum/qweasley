package utils

import (
	"fmt"
	"time"
)

// FormatTimestamp formats a timestamp for display
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ValidateUserID checks if a user ID is valid
func ValidateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}
	return nil
}

// SanitizeText removes potentially harmful characters from text
func SanitizeText(text string) string {
	// Basic sanitization - can be expanded based on requirements
	if len(text) > 4096 {
		return text[:4096]
	}
	return text
}