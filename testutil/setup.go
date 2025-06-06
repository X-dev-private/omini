// Copyright Tharsis Labs Ltd.(omini)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/omini/omini/blob/main/LICENSE)
package testutil

import (
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/version"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHeader creates a new Tendermint header for testing purposes.
func NewHeader(
	height int64,
	blockTime time.Time,
	chainID string,
	proposer sdk.ConsAddress,
	appHash,
	validatorHash []byte,
) tmproto.Header {
	return tmproto.Header{
		ChainID:         chainID,
		Height:          height,
		Time:            blockTime,
		ValidatorsHash:  validatorHash,
		AppHash:         appHash,
		ProposerAddress: proposer.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		DataHash:           tmhash.Sum([]byte("data")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
	}
}
