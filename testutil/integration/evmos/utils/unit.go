// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)
//
// This file contains all utility function that require the access to the unit
// test network and should only be used in unit tests.
package utils

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/omini/omini/v20/testutil/integration/omini/network"
	ominitypes "github.com/omini/omini/v20/types"
	erc20types "github.com/omini/omini/v20/x/erc20/types"
	inflationtypes "github.com/omini/omini/v20/x/inflation/v1/types"
)

const (
	TokenToMint = 1e18
)

// RegisterominiERC20Coins uses the UnitNetwork to register the omini token as an
// ERC20 token. The function performs all the required steps for the registration
// like registering the denom trace in the transfer keeper and minting the token
// with the bank. Returns the TokenPair or an error.
func RegisterominiERC20Coins(
	network network.UnitTestNetwork,
	tokenReceiver sdk.AccAddress,
) (erc20types.TokenPair, error) {
	bondDenom, err := network.App.StakingKeeper.BondDenom(network.GetContext())
	if err != nil {
		return erc20types.TokenPair{}, err
	}
	coin := sdk.NewCoin(ominitypes.BaseDenom, math.NewInt(TokenToMint))
	err = network.App.BankKeeper.MintCoins(
		network.GetContext(),
		inflationtypes.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}
	err = network.App.BankKeeper.SendCoinsFromModuleToAccount(
		network.GetContext(),
		inflationtypes.ModuleName,
		tokenReceiver,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	ominiMetadata, found := network.App.BankKeeper.GetDenomMetaData(network.GetContext(), ominitypes.BaseDenom)
	if !found {
		return erc20types.TokenPair{}, fmt.Errorf("expected omini denom metadata")
	}

	_, err = network.App.Erc20Keeper.RegisterERC20Extension(network.GetContext(), ominiMetadata.Base)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	ominiDenomID := network.App.Erc20Keeper.GetDenomMap(network.GetContext(), bondDenom)
	tokenPair, ok := network.App.Erc20Keeper.GetTokenPair(network.GetContext(), ominiDenomID)
	if !ok {
		return erc20types.TokenPair{}, fmt.Errorf("expected omini erc20 token pair")
	}

	return tokenPair, nil
}

// RegisterIBCERC20Coins uses the UnitNetwork to register the denomTrace as an
// ERC20 token. The function performs all the required steps for the registration
// like registering the denom trace in the transfer keeper and minting the token
// with the bank. Returns the TokenPair or an error.
func RegisterIBCERC20Coins(
	network *network.UnitTestNetwork,
	tokenReceiver sdk.AccAddress,
	denomTrace transfertypes.DenomTrace,
) (erc20types.TokenPair, error) {
	ibcDenom := denomTrace.IBCDenom()
	network.App.TransferKeeper.SetDenomTrace(network.GetContext(), denomTrace)
	ibcMetadata := banktypes.Metadata{
		Name:        "Generic IBC name",
		Symbol:      "IBC",
		Description: "Generic IBC token description",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    ibcDenom,
				Exponent: 0,
				Aliases:  []string{ibcDenom},
			},
			{
				Denom:    ibcDenom,
				Exponent: 18,
			},
		},
		Display: ibcDenom,
		Base:    ibcDenom,
	}

	coin := sdk.NewCoin(ibcMetadata.Base, math.NewInt(TokenToMint))
	err := network.App.BankKeeper.MintCoins(
		network.GetContext(),
		inflationtypes.ModuleName,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	err = network.App.BankKeeper.SendCoinsFromModuleToAccount(
		network.GetContext(),
		inflationtypes.ModuleName,
		tokenReceiver,
		sdk.NewCoins(coin),
	)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	_, err = network.App.Erc20Keeper.RegisterERC20Extension(network.GetContext(), ibcMetadata.Base)
	if err != nil {
		return erc20types.TokenPair{}, err
	}

	ibcDenomID := network.App.Erc20Keeper.GetDenomMap(
		network.GetContext(),
		denomTrace.IBCDenom(),
	)
	tokenPair, ok := network.App.Erc20Keeper.GetTokenPair(network.GetContext(), ibcDenomID)
	if !ok {
		return erc20types.TokenPair{}, fmt.Errorf("expected %s erc20 token pair", ibcDenom)
	}

	return tokenPair, nil
}
