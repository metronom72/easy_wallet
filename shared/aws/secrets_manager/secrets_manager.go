package secrets_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
	"os"
	"strings"
)

func FetchSecret(secretName string, key string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("unable to load AWS Config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("unable to get secret: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret does not exist")
	}

	if key == "" {
		return *result.SecretString, nil
	}

	var secret map[string]interface{}
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		return "", fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	if secret[key] == nil {
		return "", fmt.Errorf("secret does not exist")
	}

	value, ok := secret[key].(string)
	if !ok {
		return "", fmt.Errorf("ключ '%s' отсутствует в секрете '%s'", key, secret)
	}

	return value, nil
}

func StoreSecret(ctx context.Context, secretName string, value string) error {
	if !strings.HasPrefix(secretName, "/") {
		secretName = "/" + secretName
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Println("[ERROR] AWS_REGION environment variable is not set")
		return nil
	}

	log.Printf("[INFO] Initializing AWS Secrets Manager client in region: %s", region)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("[ERROR] Failed to load AWS config: %v", err)
		return err
	}

	secretsManagerSvc := secretsmanager.NewFromConfig(cfg)

	log.Printf("[INFO] Storing private key in AWS Secrets Manager under: %s", secretName)
	_, err = secretsManagerSvc.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(value),
	})

	if err != nil {
		log.Printf("[WARN] Secret already exists, updating instead: %s", secretName)
		_, err = secretsManagerSvc.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(value),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to update secret in AWS Secrets Manager: %v", err)
			return err
		}
		log.Printf("[SUCCESS] Secret updated in AWS Secrets Manager: %s", secretName)
	} else {
		log.Printf("[SUCCESS] Secret securely stored in AWS Secrets Manager: %s", secretName)
	}

	return nil
}
