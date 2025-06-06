// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)

package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// CoinEq returns whether two Coins are equal.
// The IsEqual() method can panic.
func CoinEq(a, b sdk.Coins) bool {
	return a.IsAllLTE(b) && b.IsAllLTE(a)
}

// Max64 returns the maximum of its inputs.
func Max64(i, j int64) int64 {
	if i > j {
		return i
	}
	return j
}

// Min64 returns the minimum of its inputs.
func Min64(i, j int64) int64 {
	if i < j {
		return i
	}
	return j
}
