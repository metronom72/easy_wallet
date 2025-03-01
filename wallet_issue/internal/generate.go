package internal

import (
	"encoding/hex"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

func GenerateWallet() (privateKeyHex, publicKeyHex string, err error) {
	log.Println("[INFO] Starting wallet generation...")

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Printf("[ERROR] Failed to generate Ethereum private key: %v", err)
		return "", "", err
	}
	log.Println("[SUCCESS] Ethereum private key generated")

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex = hex.EncodeToString(privateKeyBytes)
	log.Printf("[INFO] Private key converted to hex: %s", privateKeyHex[:10]+"... (truncated)")

	publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)
	publicKeyHex = hex.EncodeToString(publicKeyBytes)
	log.Printf("[INFO] Public key converted to hex: %s", publicKeyHex[:10]+"... (truncated)")

	log.Println("[SUCCESS] Wallet generation completed successfully")
	return privateKeyHex, publicKeyHex, nil
}
