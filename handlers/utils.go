package handlers

import (
	"crypto/rand"
	"encoding/hex"
)

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
