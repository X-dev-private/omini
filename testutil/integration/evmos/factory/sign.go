// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)
package factory

import (
	errorsmod "cosmossdk.io/errors"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/omini/omini/v20/testutil/tx"
	evmtypes "github.com/omini/omini/v20/x/evm/types"
)

// SignMsgEthereumTx signs a MsgEthereumTx with the provided private key and chainID.
func (tf *IntegrationTxFactory) SignMsgEthereumTx(privKey cryptotypes.PrivKey, msgEthereumTx evmtypes.MsgEthereumTx) (evmtypes.MsgEthereumTx, error) {
	ethChainID := tf.network.GetEIP155ChainID()
	signer := gethtypes.LatestSignerForChainID(ethChainID)
	err := msgEthereumTx.Sign(signer, tx.NewSigner(privKey))
	if err != nil {
		return evmtypes.MsgEthereumTx{}, errorsmod.Wrap(err, "failed to sign transaction")
	}
	return msgEthereumTx, nil
}
