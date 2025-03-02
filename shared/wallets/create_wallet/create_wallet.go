package create_wallet

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"os"
	"shared/aws/secrets_manager"
	"shared/encrypt"
)

type Wallet struct {
	ID        string
	PublicKey string
	SecretRef string
}

func CreateWallet(ctx context.Context, id, password, privateKey, publicKey string) (string, error) {
	tableName := os.Getenv("DYNAMO_TABLE")
	region := os.Getenv("AWS_REGION")

	if tableName == "" {
		return "", errors.New("DYNAMO_TABLE environment variable not set")
	}
	if region == "" {
		return "", errors.New("AWS_REGION environment variable not set")
	}

	log.Printf("[INFO] Creating wallet in DynamoDB table: %s (Region: %s)", tableName, region)

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return "", fmt.Errorf("error loading AWS config: %v", err)
	}

	db := dynamodb.NewFromConfig(cfg)

	encryptedId, err := encrypt.Encrypt(id, password)

	encryptedPrivateKey, err := encrypt.Encrypt(privateKey, password)
	if err != nil {
		log.Printf("Error encrypting private key: %v", err)
		return "", fmt.Errorf("error encrypting private key: %v", err)
	}

	wallet := Wallet{
		ID:        id,
		PublicKey: publicKey,
		SecretRef: "/wallet/private" + encryptedId,
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"id":         &types.AttributeValueMemberS{Value: wallet.ID},
			"public_key": &types.AttributeValueMemberS{Value: wallet.PublicKey},
			"secret_ref": &types.AttributeValueMemberS{Value: wallet.SecretRef},
		},
	})

	if err != nil {
		log.Printf("[ERROR] Failed to store wallet in DynamoDB: %v", err)
		return "", fmt.Errorf("failed to store wallet in DynamoDB: %v", err)
	}

	err = secrets_manager.StoreSecret(ctx, wallet.SecretRef, encryptedPrivateKey)
	if err != nil {
		log.Printf("Error storing secret: %v", err)
		return "", fmt.Errorf("error storing secret: %v", err)
	}

	log.Printf("[SUCCESS] Created wallet in DynamoDB: %s", wallet.ID)
	return wallet.PublicKey, nil
}
