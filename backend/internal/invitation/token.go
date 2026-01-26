package invitation

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateToken generates a secure, unique token for player invitations.
func GenerateToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Base64 encode to get a URL-safe string
	return base64.URLEncoding.EncodeToString(bytes), nil
}
