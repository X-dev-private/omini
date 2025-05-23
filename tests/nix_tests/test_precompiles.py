import re

import pytest

from .ibc_utils import (
    omini_IBC_DENOM,
    assert_ready,
    get_balance,
    get_balances,
    prepare_network,
)
from .network import omini
from .utils import (
    ADDRS,
    CONTRACTS,
    KEYS,
    amount_of,
    deploy_contract,
    get_fee,
    get_precompile_contract,
    send_transaction,
    wait_for_fn,
)


@pytest.fixture(scope="module", params=["omini", "omini-6dec", "omini-rocksdb"])
def ibc(request, tmp_path_factory):
    """
    Prepares the network.
    """
    name = "ibc-precompile"
    omini_build = request.param
    path = tmp_path_factory.mktemp(name)
    network = prepare_network(path, name, [omini_build, "chainmain"])
    yield from network


def test_ibc_transfer(ibc):
    """
    test transfer IBC precompile.
    """
    assert_ready(ibc)

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.chains["omini"].cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aomini"

    old_src_balances = get_balances(ibc.chains["omini"], src_addr)
    old_dst_balances = get_balances(ibc.chains["chainmain"], dst_addr)

    pc = get_precompile_contract(ibc.chains["omini"].w3, "ICS20I")
    omini_gas_price = ibc.chains["omini"].w3.eth.gas_price

    tx = pc.functions.transfer(
        "transfer",
        "channel-0",
        src_denom,
        amt,
        ADDRS["signer2"],
        dst_addr,
        [1, 10000000000],
        0,
        "",
    ).build_transaction(
        {
            "from": ADDRS["signer2"],
            "gasPrice": omini_gas_price,
        }
    )
    gas_estimation = ibc.chains["omini"].w3.eth.estimate_gas(tx)
    receipt = send_transaction(ibc.chains["omini"].w3, tx, KEYS["signer2"])

    assert receipt.status == 1

    # check ibc-transfer event was emitted
    transfer_event = pc.events.IBCTransfer().processReceipt(receipt)[0]
    assert transfer_event.address == "0x0000000000000000000000000000000000000802"
    assert transfer_event.event == "IBCTransfer"
    assert transfer_event.args.sender == ADDRS["signer2"]
    # TODO check if we want to keep the keccak256 hash bytes or smth better
    # assert transfer_event.args.receiver == dst_addr
    assert transfer_event.args.sourcePort == "transfer"
    assert transfer_event.args.sourceChannel == "channel-0"
    assert transfer_event.args.denom == src_denom
    assert transfer_event.args.amount == amt
    assert transfer_event.args.memo == ""

    # check gas used
    assert receipt.gasUsed == 48184

    # check gas estimation is accurate
    assert receipt.gasUsed == gas_estimation

    fee = get_fee(cli, omini_gas_price, gas_estimation, receipt.gasUsed)
    fee_denom = cli.evm_denom()

    new_dst_balances = []

    def check_balance_change():
        nonlocal new_dst_balances
        new_dst_balances = get_balances(ibc.chains["chainmain"], dst_addr)
        return amount_of(old_dst_balances, omini_IBC_DENOM) != amount_of(
            new_dst_balances, omini_IBC_DENOM
        )

    wait_for_fn("balance change", check_balance_change)
    assert amount_of(old_dst_balances, omini_IBC_DENOM) + amt == amount_of(
        new_dst_balances, omini_IBC_DENOM
    )
    new_src_balances = get_balances(ibc.chains["omini"], src_addr)
    old_src_denom_amt = amount_of(old_src_balances, src_denom)
    new_src_denom_amt = amount_of(new_src_balances, src_denom)
    if fee_denom == src_denom:
        assert old_src_denom_amt - amt - fee == new_src_denom_amt
        return

    assert old_src_denom_amt - amt == new_src_denom_amt

    old_src_fee_denom_amt = amount_of(old_src_balances, fee_denom)
    new_src_fee_denom_amt = amount_of(new_src_balances, fee_denom)
    assert old_src_fee_denom_amt - fee == new_src_fee_denom_amt


def test_ibc_transfer_invalid_packet(ibc):
    """
    test transfer IBC precompile invalid packet error.
    NOTE: it is important for this error message to not change
    because it is already stored on mainnet.
    Changing this error message is a state breaking change
    """
    assert_ready(ibc)

    # IMPORTANT: THIS ERROR MSG SHOULD NEVER CHANGE OR WILL BE A STATE BREAKING CHANGE ON MAINNET
    exp_err = "constructed packet failed basic validation: packet timeout height and packet timeout timestamp cannot both be 0: invalid packet"  # noqa: E501 # pylint: disable=line-too-long

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.chains["omini"].cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aomini"

    old_src_balance = get_balance(ibc.chains["omini"], src_addr, src_denom)

    pc = get_precompile_contract(ibc.chains["omini"].w3, "ICS20I")
    omini_gas_price = ibc.chains["omini"].w3.eth.gas_price

    try:
        tx = pc.functions.transfer(
            "transfer",
            "channel-0",
            src_denom,
            amt,
            ADDRS["signer2"],
            dst_addr,
            [0, 0],
            0,
            "",
        ).build_transaction({"from": ADDRS["signer2"], "gasPrice": omini_gas_price})
        send_transaction(ibc.chains["omini"].w3, tx, KEYS["signer2"])
    except Exception as error:
        assert error.args[0]["message"] == f"rpc error: code = Unknown desc = {exp_err}"

        new_src_balance = get_balance(ibc.chains["omini"], src_addr, src_denom)
        assert old_src_balance == new_src_balance


def test_ibc_transfer_timeout(ibc):
    """
    test transfer IBC precompile timeout packet error.
    NOTE: it is important for this error message to not change
    because it is already stored on mainnet.
    Changing this error message is a state breaking change
    """
    assert_ready(ibc)

    # IMPORTANT: THIS ERROR MSG SHOULD NEVER CHANGE OR WILL BE A STATE BREAKING CHANGE ON MAINNET
    exp_err = r"rpc error: code = Unknown desc = invalid packet timeout: current timestamp: \d+, timeout timestamp \d+: timeout elapsed"  # noqa: E501 # pylint: disable=line-too-long

    dst_addr = ibc.chains["chainmain"].cosmos_cli().address("signer2")
    amt = 1000000

    cli = ibc.chains["omini"].cosmos_cli()
    src_addr = cli.address("signer2")
    src_denom = "aomini"

    old_src_balance = get_balance(ibc.chains["omini"], src_addr, src_denom)

    pc = get_precompile_contract(ibc.chains["omini"].w3, "ICS20I")
    omini_gas_price = ibc.chains["omini"].w3.eth.gas_price

    try:
        tx = pc.functions.transfer(
            "transfer",
            "channel-0",
            src_denom,
            amt,
            ADDRS["signer2"],
            dst_addr,
            [0, 0],
            1000,
            "",
        ).build_transaction({"from": ADDRS["signer2"], "gasPrice": omini_gas_price})
        send_transaction(ibc.chains["omini"].w3, tx, KEYS["signer2"])
    except Exception as error:
        assert re.search(exp_err, error.args[0]["message"]) is not None

        new_src_balance = get_balance(ibc.chains["omini"], src_addr, src_denom)
        assert old_src_balance == new_src_balance


def test_staking(ibc):
    assert_ready(ibc)

    omini: omini = ibc.chains["omini"]
    w3 = omini.w3
    amt = 1000000
    cli = omini.cosmos_cli()
    del_addr = cli.address("signer2")
    src_denom = "aomini"
    validator_addr = cli.validators()[0]["operator_address"]

    old_src_balances = get_balances(omini, del_addr)

    pc = get_precompile_contract(w3, "StakingI")
    omini_gas_price = w3.eth.gas_price

    tx = pc.functions.delegate(ADDRS["signer2"], validator_addr, amt).build_transaction(
        {"from": ADDRS["signer2"], "gasPrice": omini_gas_price}
    )
    gas_estimation = omini.w3.eth.estimate_gas(tx)
    receipt = send_transaction(w3, tx, KEYS["signer2"])

    assert receipt.status == 1
    # check gas estimation is accurate
    assert receipt.gasUsed == gas_estimation

    fee_denom = cli.evm_denom()
    fee = get_fee(cli, omini_gas_price, gas_estimation, receipt.gasUsed)

    delegations = cli.get_delegated_amount(del_addr)["delegation_responses"]
    assert len(delegations) == 1
    assert delegations[0]["delegation"]["validator_address"] == validator_addr
    assert int(delegations[0]["balance"]["amount"]) == amt

    new_src_balances = get_balances(omini, del_addr)

    old_src_denom_amt = amount_of(old_src_balances, src_denom)
    new_src_denom_amt = amount_of(new_src_balances, src_denom)
    if fee_denom == src_denom:
        assert old_src_denom_amt - amt - fee == new_src_denom_amt
        return

    assert old_src_denom_amt - amt == new_src_denom_amt

    old_src_fee_denom_amt = amount_of(old_src_balances, fee_denom)
    new_src_fee_denom_amt = amount_of(new_src_balances, fee_denom)
    assert old_src_fee_denom_amt - fee == new_src_fee_denom_amt


def test_staking_via_sc(ibc):
    assert_ready(ibc)

    omini: omini = ibc.chains["omini"]
    w3 = omini.w3
    amt = 1000000
    cli = omini.cosmos_cli()
    del_addr = cli.address("signer1")
    src_denom = "aomini"
    validator_addr = cli.validators()[0]["operator_address"]
    fee_denom = cli.evm_denom()

    old_src_balances = get_balances(omini, del_addr)

    contract, receipt = deploy_contract(w3, CONTRACTS["StakingCaller"])
    omini_gas_price = w3.eth.gas_price

    # create grant - need to specify gas otherwise will fail with out of gas
    gas_wanted = 60000
    approve_tx = contract.functions.testApprove(
        receipt.contractAddress, ["/cosmos.staking.v1beta1.MsgDelegate"], amt
    ).build_transaction(
        {"from": ADDRS["signer1"], "gasPrice": omini_gas_price, "gas": gas_wanted}
    )

    gas_estimation = omini.w3.eth.estimate_gas(approve_tx)
    receipt = send_transaction(w3, approve_tx, KEYS["signer1"])

    assert receipt.status == 1
    # check gas estimation is accurate
    print(f"gas used: {receipt.gasUsed}")
    print(f"gas estimation: {gas_estimation}")
    # FIXME gas estimation > than gasUsed. Should be equal
    # assert receipt.gasUsed == gas_estimation

    fee1 = get_fee(cli, omini_gas_price, gas_wanted, receipt.gasUsed)

    # delegate - need to specify gas otherwise will fail with out of gas
    gas_wanted = 180000
    delegate_tx = contract.functions.testDelegate(
        ADDRS["signer1"], validator_addr, amt
    ).build_transaction(
        {"from": ADDRS["signer1"], "gasPrice": omini_gas_price, "gas": gas_wanted}
    )
    gas_estimation = omini.w3.eth.estimate_gas(delegate_tx)
    receipt = send_transaction(w3, delegate_tx, KEYS["signer1"])

    assert receipt.status == 1
    # check gas estimation is accurate
    print(f"gas used: {receipt.gasUsed}")
    print(f"gas estimation: {gas_estimation}")
    # FIXME gas estimation > than gasUsed. Should be equal
    # assert receipt.gasUsed == gas_estimation

    fee2 = get_fee(cli, omini_gas_price, gas_wanted, receipt.gasUsed)

    fees = fee1 + fee2

    delegations = cli.get_delegated_amount(del_addr)["delegation_responses"]
    assert len(delegations) == 1
    assert delegations[0]["delegation"]["validator_address"] == validator_addr
    assert int(delegations[0]["balance"]["amount"]) == amt

    new_src_balances = get_balances(omini, del_addr)

    old_src_denom_amt = amount_of(old_src_balances, src_denom)
    new_src_denom_amt = amount_of(new_src_balances, src_denom)
    if fee_denom == src_denom:
        assert old_src_denom_amt - amt - fees == new_src_denom_amt
        return

    assert old_src_denom_amt - amt == new_src_denom_amt

    old_src_fee_denom_amt = amount_of(old_src_balances, fee_denom)
    new_src_fee_denom_amt = amount_of(new_src_balances, fee_denom)
    assert old_src_fee_denom_amt - fees == new_src_fee_denom_amt
