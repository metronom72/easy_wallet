package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
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

func Verify(tokenString string) (*jwt.Token, error) {
	secret, err := GetBotToken()
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
