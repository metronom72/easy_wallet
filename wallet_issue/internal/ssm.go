package internal

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func StorePrivateKey(secretName, privateKey string) error {
	if !strings.HasPrefix(secretName, "/") {
		secretName = "/" + secretName
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Println("[ERROR] AWS_REGION environment variable is not set")
		return nil
	}

	log.Printf("[INFO] Initializing AWS Secrets Manager session in region: %s", region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create AWS session: %v", err)
		return err
	}

	secretsManagerSvc := secretsmanager.New(sess)

	log.Printf("[INFO] Storing private key in AWS Secrets Manager under: %s", secretName)
	_, err = secretsManagerSvc.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(privateKey),
	})

	if err != nil {
		log.Printf("[WARN] Secret already exists, updating instead: %s", secretName)
		_, err = secretsManagerSvc.PutSecretValue(&secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(privateKey),
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
