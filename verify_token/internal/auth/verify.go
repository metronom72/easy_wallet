package auth

import (
	"crt-mmc/shared/aws/secrets_manager"
	"errors"
	"os"
)

func GetBotToken() (string, error) {
	secretName := os.Getenv("BOT_TOKEN_SECRET_NAME")
	if secretName == "" {
		return "", errors.New("BOT_TOKEN_SECRET_NAME is not set in environment variables")
	}

	botToken, err := secrets_manager.FetchSecret(secretName, "BOT_TOKEN")
	if err != nil {
		return "", err
	}

	return botToken, nil
}
