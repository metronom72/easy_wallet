package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"log"
	"shared/encrypt"
)

func Decrypt(encryptedHex, passphrase string) (string, error) {
	log.Println("[INFO] Starting decryption process...")

	key := encrypt.Md5Hash(passphrase)

	encryptedData, err := hex.DecodeString(encryptedHex)
	if err != nil {
		log.Printf("[ERROR] Failed to decode encrypted data: %v", err)
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[ERROR] Failed to create AES cipher: %v", err)
		return "", err
	}

	nonce, ciphertext := encryptedData[:12], encryptedData[12:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR] Failed to create AES-GCM mode: %v", err)
		return "", err
	}

	decryptedData, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Printf("[ERROR] Decryption failed: %v", err)
		return "", err
	}

	log.Println("[SUCCESS] Data decrypted successfully")
	return string(decryptedData), nil
}
