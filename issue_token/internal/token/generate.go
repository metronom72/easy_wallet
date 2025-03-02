package token

import (
	"errors"
	"fmt"
	"issue_token/internal/auth"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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

func GenerateJWT(dataCheckString string) (string, error) {
	botToken, err := auth.GetBotToken()
	if err != nil {
		return "", fmt.Errorf("failed to get bot token from Secrets Manager: %v", err)
	}

	payload, err := ParseDataCheckString(dataCheckString)
	if err != nil {
		return "", fmt.Errorf("failed to parse dataCheckString: %v", err)
	}

	payload["exp"] = time.Now().Add(time.Minute * 5).Unix() // Токен действует 1 час
	payload["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))

	signedToken, err := token.SignedString([]byte(botToken))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}
