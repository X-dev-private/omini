// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)
package vesting_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/omini/omini/v20/precompiles/common"
	"github.com/omini/omini/v20/precompiles/vesting"
)

func (s *PrecompileTestSuite) TestBalances() {
	var ctx sdk.Context

	testCases := []struct {
		name        string
		malleate    func() []interface{}
		gas         uint64
		postCheck   func(data []byte)
		expError    bool
		errContains string
	}{
		{
			"fail - empty input args",
			func() []interface{} {
				return []interface{}{}
			},
			200000,
			func([]byte) {},
			true,
			fmt.Sprintf(cmn.ErrInvalidNumberOfArgs, 1, 0),
		},
		{
			"fail - invalid address",
			func() []interface{} {
				return []interface{}{
					"12asji1",
				}
			},
			200000,
			func([]byte) {},
			true,
			"invalid type for vestingAddress",
		},
		{
			"fail - account is not a vesting account",
			func() []interface{} {
				return []interface{}{
					s.keyring.GetAddr(0),
				}
			},
			200000,
			func([]byte) {},
			true,
			"is not a vesting account",
		},
		{
			"success - should return vesting account balances",
			func() []interface{} {
				s.CreateTestClawbackVestingAccount(ctx, s.keyring.GetAddr(0), toAddr)
				s.FundTestClawbackVestingAccount()
				return []interface{}{
					toAddr,
				}
			},
			200000,
			func(data []byte) {
				var out vesting.BalancesOutput
				err := s.precompile.UnpackIntoInterface(&out, vesting.BalancesMethod, data)
				s.Require().NoError(err)
				s.Require().Equal(out.Locked, lockupPeriods[0].Amount)
				s.Require().Equal(out.Unvested, lockupPeriods[0].Amount)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest(2) // reset
			ctx = s.network.GetContext()
			method := s.precompile.Methods[vesting.BalancesMethod]

			bz, err := s.precompile.Balances(ctx, &method, tc.malleate())

			if tc.expError {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(bz)
				tc.postCheck(bz)
			}
		})
	}
}
