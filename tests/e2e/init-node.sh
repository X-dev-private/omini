#!/bin/bash

CHAINID="${CHAIN_ID:-omini_9002-1}"
BASE_DENOM="aomini"
MONIKER="localtestnet"
KEYRING="test"          # remember to change to other types of keyring like 'file' in-case exposing to outside world, otherwise your balance will be wiped quickly. The keyring test does not require private key to steal tokens from you
KEYALGO="eth_secp256k1" #gitleaks:allow
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""
PRUNING="default"
#PRUNING="custom"

CHAINDIR="$HOME/.ominid"
GENESIS="$CHAINDIR/config/genesis.json"
TMP_GENESIS="$CHAINDIR/config/tmp_genesis.json"
APP_TOML="$CHAINDIR/config/app.toml"
CONFIG_TOML="$CHAINDIR/config/config.toml"

# feemarket params basefee
BASEFEE=1000000000

# myKey address 0x7cb61d4117ae31a12e393a1cfa3bac666481d02e
VAL_KEY="mykey"
VAL_MNEMONIC="gesture inject test cycle original hollow east ridge hen combine junk child bacon zero hope comfort vacuum milk pitch cage oppose unhappy lunar seat"

# user1 address 0xc6fe5d33615a1c52c08018c47e8bc53646a0e101
USER1_KEY="user1"
USER1_MNEMONIC="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"

# user2 address 0x963ebdf2e1f8db8707d05fc75bfeffba1b5bac17
USER2_KEY="user2"
USER2_MNEMONIC="maximum display century economy unlock van census kite error heart snow filter midnight usage egg venture cash kick motor survey drastic edge muffin visual"

# user3 address 0x40a0cb1C63e026A81B55EE1308586E21eec1eFa9
USER3_KEY="user3"
USER3_MNEMONIC="will wear settle write dance topic tape sea glory hotel oppose rebel client problem era video gossip glide during yard balance cancel file rose"

# user4 address 0x498B5AeC5D439b733dC2F58AB489783A23FB26dA
USER4_KEY="user4"
USER4_MNEMONIC="doll midnight silk carpet brush boring pluck office gown inquiry duck chief aim exit gain never tennis crime fragile ship cloud surface exotic patch"

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
	echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
	exit 1
}

# used to exit on first error (any non-zero exit code)
set -e

# Check ominid version to decide how to set the client configuration
# the older versions of ominid accept less arguments
sdk_version=$(ominid version --long | grep 'cosmos_sdk_version' | awk '{print $2}')
if [[ $sdk_version == *v0.4* ]]; then
	ominid config chain-id "$CHAINID"
	ominid config keyring-backend "$KEYRING"
else
	ominid config set client chain-id "$CHAINID"
	ominid config set client keyring-backend "$KEYRING"
fi

# Import keys from mnemonics
echo "$VAL_MNEMONIC" | ominid keys add "$VAL_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"

echo "$USER1_MNEMONIC" | ominid keys add "$USER1_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"
echo "$USER2_MNEMONIC" | ominid keys add "$USER2_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"
echo "$USER3_MNEMONIC" | ominid keys add "$USER3_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"
echo "$USER4_MNEMONIC" | ominid keys add "$USER4_KEY" --recover --keyring-backend "$KEYRING" --algo "$KEYALGO"

# Set moniker and chain-id for omini (Moniker can be anything, chain-id must be an integer)
ominid init "$MONIKER" --chain-id "$CHAINID"

# Change parameter token denominations to $BASE_DENOM
jq --arg base_denom "$BASE_DENOM" '.app_state["staking"]["params"]["bond_denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg base_denom "$BASE_DENOM" '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
# When upgrade to cosmos-sdk v0.47, use gov.params to edit the deposit params
jq --arg base_denom "$BASE_DENOM" '.app_state["gov"]["params"]["min_deposit"][0]["denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
# jq '.app_state["evm"]["params"]["evm_denom"]="$BASE_DENOM"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq --arg base_denom "$BASE_DENOM" '.app_state["inflation"]["params"]["mint_denom"]=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set gov proposing && voting period
jq '.app_state.gov.deposit_params.max_deposit_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
jq '.app_state.gov.voting_params.voting_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
sed -i.bak 's/"expedited_voting_period": "86400s"/"expedited_voting_period": "5s"/g' "$GENESIS"

# When upgrade to cosmos-sdk v0.47, use gov.params to edit the deposit params
# check if the 'params' field exists in the genesis file
if jq '.app_state.gov.params != null' "$GENESIS" | grep -q "true"; then
	jq --arg base_denom "$BASE_DENOM" '.app_state.gov.params.min_deposit[0].denom=$base_denom' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state.gov.params.max_deposit_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
	jq '.app_state.gov.params.voting_period="10s"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
fi

# Set gas limit in genesis
jq '.consensus_params.block.max_gas="10000000"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set claims start time
current_date=$(date -u +"%Y-%m-%dT%TZ")
jq -r --arg current_date "$current_date" '.app_state.claims.params.airdrop_start_time=$current_date' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# Set base fee in genesis
jq '.app_state["feemarket"]["params"]["base_fee"]="'${BASEFEE}'"' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# disable produce empty block
sed -i.bak 's/create_empty_blocks = true/create_empty_blocks = false/g' "$CONFIG_TOML"

# Allocate genesis accounts (cosmos formatted addresses)
ominid add-genesis-account "$(ominid keys show "$VAL_KEY" -a --keyring-backend "$KEYRING")" 100000000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING"
ominid add-genesis-account "$(ominid keys show "$USER1_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING"
ominid add-genesis-account "$(ominid keys show "$USER2_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING"
ominid add-genesis-account "$(ominid keys show "$USER3_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING"
ominid add-genesis-account "$(ominid keys show "$USER4_KEY" -a --keyring-backend "$KEYRING")" 1000000000000000000000$BASE_DENOM --keyring-backend "$KEYRING"

# Update total supply with claim values
# Bc is required to add this big numbers
# total_supply=$(bc <<< "$validators_supply")
total_supply=100004000000000000000000000
jq -r --arg total_supply "$total_supply" '.app_state.bank.supply[0].amount=$total_supply' "$GENESIS" >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

# set custom pruning settings
if [ "$PRUNING" = "custom" ]; then
	sed -i.bak 's/pruning = "default"/pruning = "custom"/g' "$APP_TOML"
	sed -i.bak 's/pruning-keep-recent = "0"/pruning-keep-recent = "2"/g' "$APP_TOML"
	sed -i.bak 's/pruning-interval = "0"/pruning-interval = "10"/g' "$APP_TOML"
fi

# make sure the localhost IP is 0.0.0.0
sed -i.bak 's/localhost/0.0.0.0/g' "$CONFIG_TOML"
sed -i.bak 's/127.0.0.1/0.0.0.0/g' "$APP_TOML"

# use timeout_commit 1s to make test faster
sed -i.bak 's/timeout_commit = "3s"/timeout_commit = "1s"/g' "$CONFIG_TOML"

# Sign genesis transaction
ominid gentx "$VAL_KEY" 1000000000000000000000$BASE_DENOM --gas-prices ${BASEFEE}$BASE_DENOM --keyring-backend "$KEYRING" --chain-id "$CHAINID"
## In case you want to create multiple validators at genesis
## 1. Back to `ominid keys add` step, init more keys
## 2. Back to `ominid add-genesis-account` step, add balance for those
## 3. Clone this ~/.ominid home directory into some others, let's say `~/.clonedominid`
## 4. Run `gentx` in each of those folders
## 5. Copy the `gentx-*` folders under `~/.clonedominid/config/gentx/` folders into the original `~/.ominid/config/gentx`

# Enable the APIs for the tests to be successful
sed -i.bak 's/enable = false/enable = true/g' "$APP_TOML"
# Don't enable Rosetta API by default
grep -q -F '[rosetta]' "$APP_TOML" && sed -i.bak '/\[rosetta\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
# Don't enable memiavl by default
grep -q -F '[memiavl]' "$APP_TOML" && sed -i.bak '/\[memiavl\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"
# Don't enable versionDB by default
grep -q -F '[versiondb]' "$APP_TOML" && sed -i.bak '/\[versiondb\]/,/^\[/ s/enable = true/enable = false/' "$APP_TOML"

# Collect genesis tx
ominid collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
ominid validate-genesis

# Start the node
ominid start "$TRACE" \
	--log_level $LOGLEVEL \
	--minimum-gas-prices=0.0001$BASE_DENOM \
	--json-rpc.api eth,txpool,personal,net,debug,web3 \
	--chain-id "$CHAINID"
