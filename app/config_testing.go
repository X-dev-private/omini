// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)

//go:build test
// +build test

package app

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/omini/omini/v20/x/evm/types"
)

// ominiOptionsFn defines a function type for setting app options specifically for
// the omini app. The function should receive the chainID and return an error if
// any.
type ominiOptionsFn func(string) error

// NoOpominiOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpominiOptions(_ string) error {
	return nil
}

// ominiAppOptions allows to setup the global configuration
// for the omini chain.
func ominiAppOptions(chainID string) error {
	id := strings.Split(chainID, "-")[0]
	coinInfo, found := evmtypes.ChainsCoinInfo[id]
	if !found {
		return fmt.Errorf("unknown chain id: %s", chainID)
	}

	// set the base denom considering if its mainnet or testnet
	if err := setBaseDenom(coinInfo); err != nil {
		return err
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	configurator := evmtypes.NewEVMConfigurator()
	// reset configuration to set the new one
	configurator.ResetTestConfig()
	err = configurator.
		WithExtendedEips(ominiActivators).
		WithChainConfig(ethCfg).
		WithEVMCoinInfo(baseDenom, uint8(coinInfo.Decimals)).
		Configure()
	if err != nil {
		return err
	}

	return nil
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain. The function registers different values based on
// the EvmCoinInfo to allow different configurations in mainnet and testnet.
func setBaseDenom(ci evmtypes.EvmCoinInfo) (err error) {
	// Defer setting the base denom, and capture any potential error from it.
	// So when failing because the denom was already registered, we ignore it and set
	// the corresponding denom to be base denom
	defer func() {
		err = sdk.SetBaseDenom(ci.Denom)
	}()

	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}

	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() units are lower than the current base denom's units.
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
