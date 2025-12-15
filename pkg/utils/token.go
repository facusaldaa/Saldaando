package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateInviteToken generates a secure random invitation token
func GenerateInviteToken() (string, error) {
	bytes := make([]byte, 16) // 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// FormatInviteToken formats an invite token for display
func FormatInviteToken(token string) string {
	// Format as: XXXX-XXXX-XXXX-XXXX for readability
	if len(token) >= 32 {
		return fmt.Sprintf("%s-%s-%s-%s", 
			token[0:8], token[8:16], token[16:24], token[24:32])
	}
	return token
}

// ParseInviteToken removes dashes from formatted token
func ParseInviteToken(formatted string) string {
	// Remove dashes
	result := ""
	for _, char := range formatted {
		if char != '-' {
			result += string(char)
		}
	}
	return result
}

