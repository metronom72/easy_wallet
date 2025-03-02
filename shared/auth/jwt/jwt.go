package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"shared/secrets"
	"strings"
	"time"
)

func GenerateJWT(data map[string]interface{}) (string, error) {
	botToken, err := secrets.GetBotToken()
	if err != nil {
		return "", fmt.Errorf("failed to get bot token from Secrets Manager: %v", err)
	}

	data["exp"] = time.Now().Add(time.Minute * 5).Unix() // Токен действует 1 час
	data["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(data))

	signedToken, err := token.SignedString([]byte(botToken))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return signedToken, nil
}

func Verify(tokenString string) (*jwt.Token, error) {
	if !isValidJWTFormat(tokenString) {
		return nil, fmt.Errorf("malformed token [%v]: incorrect JWT format", tokenString)
	}

	secret, err := secrets.GetBotToken()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bot token: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if int64(exp) < time.Now().Unix() {
				return nil, errors.New("token is expired")
			}
		}
	}

	return token, nil
}

func isValidJWTFormat(token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if _, err := base64.RawURLEncoding.DecodeString(part); err != nil {
			return false
		}
	}
	return true
}
