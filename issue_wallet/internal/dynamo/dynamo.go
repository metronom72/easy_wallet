package dynamo

import (
	"context"
	"shared/wallets/create_wallet"
	"shared/wallets/fetch_wallet"
)

func StoreWallet(id, password, privateKey, publicKey string) (string, error) {
	ctx := context.TODO()

	publicKey, err := fetch_wallet.FetchWallet(ctx, id, password)
	if err == nil {
		return publicKey, nil
	}

	return create_wallet.CreateWallet(ctx, id, password, privateKey, publicKey)
}
