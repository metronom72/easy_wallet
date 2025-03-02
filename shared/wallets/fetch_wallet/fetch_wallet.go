package fetch_wallet

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"os"
	"shared/aws/secrets_manager"
	"shared/decrypt"
)

func FetchWallet(ctx context.Context, id, password string) (string, error) {
	tableName := os.Getenv("DYNAMO_TABLE")
	region := os.Getenv("AWS_REGION")

	if tableName == "" {
		return "", fmt.Errorf("DYNAMO_TABLE environment variable is not set")
	}
	if region == "" {
		return "", fmt.Errorf("AWS_REGION environment variable is not set")
	}

	log.Printf("[INFO] Fetching wallet from DynamoDB table: %s (Region: %s)", tableName, region)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("[ERROR] Failed to load AWS config: %v", err)
		return "", err
	}

	db := dynamodb.NewFromConfig(cfg)

	result, err := db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		log.Printf("[ERROR] Failed to fetch wallet: %v", err)
		return "", err
	}

	if result.Item == nil {
		log.Println("[ERROR] Wallet not found")
		return "", fmt.Errorf("wallet not found")
	}

	storedPublicKey := result.Item["public_key"].(*types.AttributeValueMemberS).Value
	storedSecretRef := result.Item["secret_ref"].(*types.AttributeValueMemberS).Value

	encryptedPrivateKey, err := secrets_manager.FetchSecret(storedSecretRef, "")
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve encrypted private key: %v", err)
		return "", err
	}

	decryptedPrivateKey, err := decrypt.Decrypt(encryptedPrivateKey, password)
	if err != nil {
		log.Printf("[ERROR] Failed to decrypt private key: %v", err)
		return "", fmt.Errorf("wallet verification failed")
	}

	if decryptedPrivateKey != "" {
		log.Println("[SUCCESS] Wallet verified, returning public key")
		return storedPublicKey, nil
	}

	log.Println("[ERROR] Wallet verification failed")
	return "", fmt.Errorf("wallet verification failed")
}
