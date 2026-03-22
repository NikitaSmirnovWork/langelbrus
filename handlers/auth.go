package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// TelegramAuth validates the initData from Telegram WebApp and extracts user_id.
// In dev mode (no BOT_TOKEN set), it falls back to X-User-ID header.
func TelegramAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		botToken := os.Getenv("BOT_TOKEN")

		// Dev mode — no token set
		if botToken == "" {
			userID := c.GetHeader("X-User-ID")
			if userID == "" {
				userID = "dev_user"
			}
			c.Set("user_id", userID)
			c.Next()
			return
		}

		initData := c.GetHeader("X-Telegram-Init-Data")
		if initData == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Telegram init data"})
			c.Abort()
			return
		}

		userID, err := validateTelegramInitData(initData, botToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Telegram auth"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func validateTelegramInitData(initData, botToken string) (string, error) {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return "", err
	}

	hash := params.Get("hash")
	if hash == "" {
		return "", fmt.Errorf("missing hash")
	}

	// Build check string
	var keys []string
	for k := range params {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params.Get(k))
	}
	checkString := strings.Join(parts, "\n")

	// HMAC-SHA256
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))

	mac := hmac.New(sha256.New, secretKey.Sum(nil))
	mac.Write([]byte(checkString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if expectedHash != hash {
		return "", fmt.Errorf("hash mismatch")
	}

	// Extract user id from user JSON field
	userJSON := params.Get("user")
	if userJSON == "" {
		return "", fmt.Errorf("missing user field")
	}

	// Simple extraction without full JSON parse
	idStart := strings.Index(userJSON, `"id":`)
	if idStart == -1 {
		return "", fmt.Errorf("missing id in user")
	}
	idStart += 5
	rest := strings.TrimSpace(userJSON[idStart:])
	end := strings.IndexAny(rest, ",}")
	if end == -1 {
		return "", fmt.Errorf("malformed user id")
	}
	userID := strings.TrimSpace(rest[:end])

	return userID, nil
}
