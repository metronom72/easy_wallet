package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"shared/secrets"
	"strings"
)

func VerifyTelegramAuth(dataCheckString string, hash string) (bool, error) {
	botToken, err := secrets.GetBotToken()
	if err != nil {
		return false, fmt.Errorf("failed to get bot token from Secrets Manager: %v", err)
	}

	if dataCheckString == "" {
		return false, errors.New("dataCheckString is empty")
	}

	if hash == "" {
		return false, errors.New("hash is empty")
	}

	hashedToken := sha256.Sum256([]byte(botToken))
	secretKey := hashedToken[:]

	mac := hmac.New(sha256.New, secretKey)
	_, err = mac.Write([]byte(dataCheckString))
	if err != nil {
		return false, fmt.Errorf("failed to compute HMAC: %v", err)
	}
	expectedHash := mac.Sum(nil)

	expectedHashHex := hex.EncodeToString(expectedHash)

	if expectedHashHex != hash {
		return false, errors.New("hash verification failed")
	}
	return true, nil
}

func ParseDataCheckString(dataCheckString string) (map[string]interface{}, error) {
	if dataCheckString == "" {
		return nil, errors.New("dataCheckString is empty")
	}

	parsedData := make(map[string]interface{})

	lines := strings.Split(dataCheckString, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid dataCheckString format: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		parsedData[key] = value
	}

	return parsedData, nil
}
