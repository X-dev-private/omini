from .utils import get_current_height, get_scaling_factor, supervisorctl, wait_for_block


def test_block_cmd(omini_cluster):
    """
    - start 2 omini nodes
    - wait for a certain height
    - stop the node1
    - use the 'block' cli cmd
    - restart omini node1
    """

    # wait for specific height
    node1 = omini_cluster.cosmos_cli(1)
    current_height = get_current_height(node1)

    last_block = current_height + 2
    wait_for_block(node1, last_block)

    # stop node1
    supervisorctl(
        omini_cluster.base_dir / "../tasks.ini",
        "stop",
        f"{omini_cluster.cosmos_cli().chain_id}-node1",
    )

    # use 'block' CLI cmd in node1
    test_cases = [
        {
            "name": "success - get latest block",
            "flags": [],
            "exp_out": f'"last_commit":{{"height":{last_block - 1}',
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "success - get block #2",
            "flags": ["--height", 2],
            "exp_out": '"height":2',
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "fail - get inexistent block",
            "flags": ["--height", last_block + 10],
            "exp_out": None,
            "exp_err": True,
            "err_msg": f"invalid height, the latest height found in the db is {last_block}, "
            f"and you asked for {last_block + 10}",
        },
    ]
    for tc in test_cases:
        try:
            output = node1.raw("block", *tc["flags"], home=node1.data_dir)
            assert tc["exp_out"] in output.decode()
        except Exception as err:
            if tc["exp_err"] is True:
                assert tc["err_msg"] in err.args[0]
                continue

            print(f"Unexpected {err=}, {type(err)=}")
            raise

    # start node1 again
    supervisorctl(
        omini_cluster.base_dir / "../tasks.ini",
        "start",
        f"{omini_cluster.cosmos_cli().chain_id}-node1",
    )
    # check if chain continues alright
    wait_for_block(node1, last_block + 3)


def test_tx_flags(omini_cluster):
    """
    Tests the expected responses for common fee and gas related CLI flags.
    """

    node = omini_cluster.cosmos_cli(0)
    current_height = get_current_height(node)
    wait_for_block(node, current_height + 1)
    fee_denom = node.evm_denom()
    scale_factor = get_scaling_factor(node)

    test_cases = [
        {
            "name": "fail - invalid flags combination (gas-prices & fees)",
            "flags": {
                "fees": f"{int(5000000000 / scale_factor)}{fee_denom}",
                "gas_prices": f"{50000/ scale_factor}{fee_denom}",
            },
            "exp_err": True,
            "err_msg": "cannot provide both fees and gas prices",
        },
        {
            "name": "fail - no fees & insufficient gas",
            "flags": {"gas": 50000, "gas_prices": None},
            "exp_err": True,
            "err_msg": "gas prices too low",
        },
        {
            "name": "fail - insufficient fees",
            "flags": {
                "fees": f"{int(10/scale_factor)}{fee_denom}",
                "gas": 50000,
                "gas_prices": None,
            },
            "exp_err": True,
            "err_msg": "insufficient fee",
        },
        {
            "name": "fail - insufficient gas",
            "flags": {
                "fees": f"{int(500000000000/scale_factor)}{fee_denom}",
                "gas": 1,
                "gas_prices": None,
            },
            "exp_err": True,
            "err_msg": "out of gas",
        },
        {
            "name": "success - defined fees & gas",
            "flags": {
                "fees": f"{int(10000000000000000000/scale_factor)}{fee_denom}",
                "gas": 1500000,
                "gas_prices": None,
            },
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "success - using gas & gas-prices",
            "flags": {
                "gas_prices": f"{100000000000/scale_factor}{fee_denom}",
                "gas": 1500000,
            },
            "exp_err": False,
            "err_msg": None,
        },
        {
            "name": "success - using gas 'auto' and specific fees",
            "flags": {
                "gas": "auto",
                "fees": f"{int(10000000000000000000/scale_factor)}{fee_denom}",
                "gas_prices": None,
            },
            "exp_err": False,
            "err_msg": None,
        },
    ]

    for tc in test_cases:
        try:
            res = node.transfer(
                "signer1",
                "omini10jmp6sgh4cc6zt3e8gw05wavvejgr5pwjnpcky",
                f"{int(100000000000000/scale_factor)}{fee_denom}",
                False,
                **tc["flags"],
            )
            if not tc["exp_err"]:
                assert res["code"] == 0, (
                    tc["name"] + ". expected tx to be successful " + res["raw_log"]
                )
                # wait for block to update nonce
                current_height = get_current_height(node)
                wait_for_block(node, current_height + 2)
            else:
                assert res["code"] != 0, tc["name"] + ". expected tx to fail"
                assert tc["err_msg"] in res["raw_log"]
        except Exception as err:
            if tc["exp_err"] is True:
                assert tc["err_msg"] in err.args[0], (
                    tc["name"]
                    + ". expected different error to be found. got "
                    + err.args[0]
                )
                continue

            print(f"Unexpected {err=}, {type(err)=}")
            raise
