package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"internal/db"
	"internal/models"
	"internal/repository"
	"log"
	"os"
	"sort"
	"strings"
)

type ProviderEnum string

const (
	ProviderTelegram ProviderEnum = "telegram"
)

type Auth struct {
	Provider ProviderEnum           `json:"provider" binding:"required"`
	Data     map[string]interface{} `json:"data"`
	UserData models.User            `json:"user_data"`
}

func (a *Auth) Authorize(ctx context.Context) (*models.User, error) {
	switch result := a.Provider; result {
	case ProviderTelegram:
		log.Println("[INFO] Authorizing telegram provider")
		authorizedUser, err := a.authorizeTelegram()
		if err != nil {
			return nil, err
		}

		conn, ok := db.GetDBFromContext(ctx)
		if !ok {
			return nil, errors.New("could not get db from context")
		}

		repo := repository.NewUserRepository(conn)

		user, err := repo.FindByExternalId(authorizedUser["id"].(string), "telegram")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user, err = repo.Create(&models.User{
				ExternalID: authorizedUser["id"].(string),
				Username:   authorizedUser["username"].(string),
				Provider:   "telegram",
			})
			if err != nil {
				return nil, err
			}
		}

		return user, nil
	default:
		return nil, errors.New("unsupported authentication provider")
	}
}

func (a *Auth) authorizeTelegram() (map[string]interface{}, error) {
	if a.Provider != ProviderTelegram {
		return nil, errors.New("unsupported authentication provider")
	}

	receivedHash, ok := a.Data["hash"].(string)
	if !ok || receivedHash == "" {
		return nil, errors.New("hash is missing in authentication data")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, errors.New("bot token is missing in environment variables")
	}

	verifiedData := make(map[string]interface{})
	for k, v := range a.Data {
		if k != "hash" {
			verifiedData[k] = v
		}
	}

	var dataPairs []string
	for key, value := range verifiedData {
		dataPairs = append(dataPairs, fmt.Sprintf("%s=%v", key, value))
	}
	sort.Strings(dataPairs)
	dataCheckString := strings.Join(dataPairs, "\n")

	secretKey := sha256.Sum256([]byte(botToken))
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	if expectedHash != receivedHash {
		return nil, errors.New("verification failed: hash mismatch")
	}

	return verifiedData, nil
}
