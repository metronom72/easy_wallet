package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"os"
)

type BotSecret struct {
	BOT_TOKEN string `json:"BOT_TOKEN"`
}

func GetBotToken() (string, error) {
	secretName := os.Getenv("BOT_TOKEN_SECRET_NAME")
	if secretName == "" {
		return "", errors.New("BOT_TOKEN_SECRET_NAME is not set in environment variables")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("unable to load AWS config: %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %v", err)
	}

	if result.SecretString == nil {
		return "", errors.New("secret value is empty")
	}

	var botSecret BotSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &botSecret); err != nil {
		return "", fmt.Errorf("failed to parse secret JSON: %v", err)
	}

	if botSecret.BOT_TOKEN == "" {
		return "", errors.New("BOT_TOKEN is missing in secret")
	}

	return botSecret.BOT_TOKEN, nil
}

func Verify(dataCheckString string, hash string) (bool, error) {
	botToken, err := GetBotToken()
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
