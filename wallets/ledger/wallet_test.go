// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)

package ledger_test

import (
	"crypto/ecdsa"
	"errors"

	gethaccounts "github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/omini/omini/v20/ethereum/eip712"
	"github.com/omini/omini/v20/wallets/accounts"
	"github.com/omini/omini/v20/wallets/ledger/mocks"
)

func RegisterDerive(mockWallet *mocks.Wallet, addr common.Address, publicKey *ecdsa.PublicKey) {
	mockWallet.On("Derive", gethaccounts.DefaultBaseDerivationPath, true).
		Return(accounts.Account{Address: addr, PublicKey: publicKey}, nil)
}

func RegisterDeriveError(mockWallet *mocks.Wallet) {
	mockWallet.On("Derive", gethaccounts.DefaultBaseDerivationPath, true).
		Return(accounts.Account{}, errors.New("unable to derive Ledger address, please open the Ethereum app and retry"))
}

func RegisterOpen(mockWallet *mocks.Wallet) {
	mockWallet.On("Open", "").
		Return(nil)
}

func RegisterClose(mockWallet *mocks.Wallet) {
	mockWallet.On("Close").
		Return(nil)
}

func RegisterSignTypedData(mockWallet *mocks.Wallet, account accounts.Account, typedDataBz []byte) {
	typedData, _ := eip712.GetEIP712TypedDataForMsg(typedDataBz)
	mockWallet.On("SignTypedData", account, typedData).
		Return([]byte{}, nil)
}

func RegisterSignTypedDataError(mockWallet *mocks.Wallet, account accounts.Account, typedDataBz []byte) {
	typedData, _ := eip712.GetEIP712TypedDataForMsg(typedDataBz)
	mockWallet.On("SignTypedData", account, typedData).
		Return([]byte{}, errors.New("error generating signature, please retry"))
}
