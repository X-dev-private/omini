// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)
package grpc

import (
	"context"
	"errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/omini/omini/v20/x/evm/core/vm"
	evmtypes "github.com/omini/omini/v20/x/evm/types"
)

// GetEvmAccount returns the EVM account for the given address.
func (gqh *IntegrationHandler) GetEvmAccount(address common.Address) (*evmtypes.QueryAccountResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Account(context.Background(), &evmtypes.QueryAccountRequest{
		Address: address.String(),
	})
}

// EstimateGas returns the estimated gas for the given call args.
func (gqh *IntegrationHandler) EstimateGas(args []byte, gasCap uint64) (*evmtypes.EstimateGasResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	res, err := evmClient.EstimateGas(context.Background(), &evmtypes.EthCallRequest{
		Args:   args,
		GasCap: gasCap,
	})
	if err != nil {
		return nil, err
	}

	// handle case where there's a revert related error
	if res.Failed() {
		if (res.VmError != vm.ErrExecutionReverted.Error()) || len(res.Ret) == 0 {
			return nil, errors.New(res.VmError)
		}
		return nil, evmtypes.NewExecErrorWithReason(res.Ret)
	}

	return res, err
}

// GetEvmParams returns the EVM module params.
func (gqh *IntegrationHandler) GetEvmParams() (*evmtypes.QueryParamsResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Params(context.Background(), &evmtypes.QueryParamsRequest{})
}

// GetEvmParams returns the EVM module params.
func (gqh *IntegrationHandler) GetEvmBaseFee() (*evmtypes.QueryBaseFeeResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.BaseFee(context.Background(), &evmtypes.QueryBaseFeeRequest{})
}

// GetBalanceFromEVM returns the balance for the given address.
func (gqh *IntegrationHandler) GetBalanceFromEVM(address sdktypes.AccAddress) (*evmtypes.QueryBalanceResponse, error) {
	evmClient := gqh.network.GetEvmClient()
	return evmClient.Balance(context.Background(), &evmtypes.QueryBalanceRequest{
		Address: common.BytesToAddress(address).Hex(),
	})
}
