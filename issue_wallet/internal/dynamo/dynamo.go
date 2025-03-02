package dynamo

import (
	"fmt"
	"github.com/metronom72/crt_mmc/wallet_issue/internal/encrypt"
	"github.com/metronom72/crt_mmc/wallet_issue/internal/ssm"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func StoreWallet(id, password, privateKey, publicKey string) error {
	tableName := os.Getenv("DYNAMO_TABLE")
	region := os.Getenv("AWS_REGION")

	if tableName == "" {
		return fmt.Errorf("[ERROR] DYNAMO_TABLE environment variable is not set")
	}
	if region == "" {
		return fmt.Errorf("[ERROR] AWS_REGION environment variable is not set")
	}

	log.Printf("[INFO] Storing wallet in DynamoDB table: %s (Region: %s)", tableName, region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create AWS session: %v", err)
		return err
	}
	db := dynamodb.New(sess)

	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	})
	if err != nil {
		log.Printf("[ERROR] Failed to check existing wallet: %v", err)
		return err
	}
	if result.Item != nil {
		log.Printf("[ERROR] Wallet with ID %s already exists", id)
		return fmt.Errorf("wallet already exists")
	}

	encryptedID, err := encrypt.Encrypt(id, password)
	if err != nil {
		log.Printf("[ERROR] Failed to encrypt ID: %v", err)
		return err
	}

	encryptedPrivateKey, err := encrypt.Encrypt(privateKey, password)
	if err != nil {
		log.Printf("[ERROR] Failed to encrypt private key: %v", err)
		return err
	}

	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]*dynamodb.AttributeValue{
			"id":         {S: aws.String(id)},
			"public_key": {S: aws.String(publicKey)},
			"secret_ref": {S: aws.String("/wallets/private/" + encryptedID)},
		},
	})
	if err != nil {
		log.Printf("[ERROR] Failed to store wallet in DynamoDB: %v", err)
		return err
	}

	log.Println("[SUCCESS] Wallet stored in DynamoDB successfully")

	err = ssm.StorePrivateKey("/wallets/private/"+encryptedID, encryptedPrivateKey)
	if err != nil {
		log.Printf("[ERROR] Failed to store encrypted private key in Secrets Manager: %v", err)
		return err
	}

	log.Println("[SUCCESS] Encrypted private key securely stored in Secrets Manager")
	return nil
}
