package evm_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/omini/omini/v20/app/ante/evm"
	"github.com/omini/omini/v20/app/ante/testutils"
	"github.com/omini/omini/v20/testutil"
	utiltx "github.com/omini/omini/v20/testutil/tx"
	"github.com/omini/omini/v20/types"
	evmtypes "github.com/omini/omini/v20/x/evm/types"
)

func (suite *AnteTestSuite) TestGasWantedDecorator() {
	suite.WithFeemarketEnabled(true)
	suite.SetupTest()
	ctx := suite.GetNetwork().GetContext()
	dec := evm.NewGasWantedDecorator(suite.GetNetwork().App.EvmKeeper, suite.GetNetwork().App.FeeMarketKeeper)
	from, fromPrivKey := utiltx.NewAddrKey()
	to := utiltx.GenerateAddress()

	baseDenom := evmtypes.GetEVMCoinDenom()

	testCases := []struct {
		name              string
		expectedGasWanted uint64
		malleate          func() sdk.Tx
		expPass           bool
	}{
		{
			"Cosmos Tx",
			testutils.TestGasLimit,
			func() sdk.Tx {
				testMsg := banktypes.MsgSend{
					FromAddress: "omini1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "omini1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: baseDenom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), baseDenom, &testMsg)
				return txBuilder.GetTx()
			},
			true,
		},
		{
			"Ethereum Legacy Tx",
			testutils.TestGasLimit,
			func() sdk.Tx {
				txArgs := evmtypes.EvmTxArgs{
					To:       &to,
					GasPrice: big.NewInt(0),
					GasLimit: testutils.TestGasLimit,
				}
				return suite.CreateTxBuilder(fromPrivKey, txArgs).GetTx()
			},
			true,
		},
		{
			"Ethereum Access List Tx",
			testutils.TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				txArgs := evmtypes.EvmTxArgs{
					To:       &to,
					GasPrice: big.NewInt(0),
					GasLimit: testutils.TestGasLimit,
					Accesses: &emptyAccessList,
				}
				return suite.CreateTxBuilder(fromPrivKey, txArgs).GetTx()
			},
			true,
		},
		{
			"Ethereum Dynamic Fee Tx (EIP1559)",
			testutils.TestGasLimit,
			func() sdk.Tx {
				emptyAccessList := ethtypes.AccessList{}
				txArgs := evmtypes.EvmTxArgs{
					To:        &to,
					GasPrice:  big.NewInt(0),
					GasFeeCap: big.NewInt(100),
					GasLimit:  testutils.TestGasLimit,
					GasTipCap: big.NewInt(50),
					Accesses:  &emptyAccessList,
				}
				return suite.CreateTxBuilder(fromPrivKey, txArgs).GetTx()
			},
			true,
		},
		{
			"EIP712 message",
			200000,
			func() sdk.Tx {
				amount := sdk.NewCoins(sdk.NewCoin(baseDenom, sdkmath.NewInt(20)))
				gas := uint64(200000)
				acc := suite.GetNetwork().App.AccountKeeper.NewAccountWithAddress(ctx, from.Bytes())
				suite.Require().NoError(acc.SetSequence(1))
				suite.GetNetwork().App.AccountKeeper.SetAccount(ctx, acc)
				builder, err := suite.CreateTestEIP712TxBuilderMsgSend(acc.GetAddress(), fromPrivKey, ctx.ChainID(), gas, amount)
				suite.Require().NoError(err)
				return builder.GetTx()
			},
			true,
		},
		{
			"Cosmos Tx - gasWanted > max block gas",
			testutils.TestGasLimit,
			func() sdk.Tx {
				testMsg := banktypes.MsgSend{
					FromAddress: "omini1x8fhpj9nmhqk8z9kpgjt95ck2xwyue0ptzkucp",
					ToAddress:   "omini1dx67l23hz9l0k9hcher8xz04uj7wf3yu26l2yn",
					Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(10), Denom: baseDenom}},
				}
				txBuilder := suite.CreateTestCosmosTxBuilder(sdkmath.NewInt(10), baseDenom, &testMsg)
				limit := types.BlockGasLimit(ctx)
				txBuilder.SetGasLimit(limit + 5)
				return txBuilder.GetTx()
			},
			false,
		},
	}

	// cumulative gas wanted from all test transactions in the same block
	var expectedGasWanted uint64

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := dec.AnteHandle(ctx, tc.malleate(), false, testutil.NextFn)
			if tc.expPass {
				suite.Require().NoError(err)

				gasWanted := suite.GetNetwork().App.FeeMarketKeeper.GetTransientGasWanted(ctx)
				expectedGasWanted += tc.expectedGasWanted
				suite.Require().Equal(expectedGasWanted, gasWanted)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
