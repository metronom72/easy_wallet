package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
)

func md5Hash(passphrase string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(passphrase))
	hash := hasher.Sum(nil)
	key := make([]byte, 32)
	copy(key, hash)
	return key
}

func Encrypt(data, passphrase string) (string, error) {
	log.Println("[INFO] Starting encryption process...")

	key := md5Hash(passphrase)
	log.Println("[INFO] Encryption key derived")

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[ERROR] Error creating AES cipher: %v", err)
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("[ERROR] Error generating nonce: %v", err)
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ERROR] Error creating AES-GCM mode: %v", err)
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(data), nil)
	encryptedHex := hex.EncodeToString(ciphertext)
	log.Println("[SUCCESS] Data encrypted successfully")

	return encryptedHex, nil
}
