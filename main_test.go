package main

import (
	"testing"

	"gopkg.in/h2non/gock.v1"
)

var (
	account = `
{
  "accounts": [
    {
      "accountIndex": "hez:ETH:1276",
      "balance": "949407923216206876",
      "bjj": "hez:0xfddace21457376b0952ccd19ce66b854fdd7c6e45905b0a0a75747c87d41719a",
      "hezEthereumAddress": "hez:0x9aC7Fdc4930e7798f9a4e014AAc0544e19b8AcE0",
      "itemId": 1045,
      "nonce": 1,
      "token": {
        "USD": 1592.32,
        "decimals": 18,
        "ethereumAddress": "0x0000000000000000000000000000000000000000",
        "ethereumBlockNum": 0,
        "fiatUpdate": "2021-03-23T10:55:20.984541Z",
        "id": 0,
        "itemId": 1,
        "name": "Ether",
        "symbol": "ETH"
      }
    }
  ],
  "pendingItems": 0
}
`
	lastBatch = `
{
  "batches": [
    {
      "itemId": 645,
      "batchNum": 625,
      "ethereumBlockNum": 8291847,
      "ethereumBlockHash": "0xb08f7d5badb05fddaa296ef6e17b89fac58146deea14ed99cdd071870e782ba7",
      "timestamp": "2021-03-24T18:38:23Z",
      "forgerAddr": "0x4fc28cd8d35b6fd644e5c1822d67609c11e137f2",
      "collectedFees": {
        "0": "781250000000000"
      },
      "historicTotalCollectedFeesUSD": 1.3165781250000002,
      "stateRoot": "16766552327640891364576382590404178978951105042545584100646269436653682377639",
      "numAccounts": 6,
      "exitRoot": "0",
      "forgeL1TransactionsNum": 624,
      "slotNum": 896,
      "forgedTransactions": 7
    }
  ],
  "pendingItems": 624
}
`
	txHistory = `
{
  "transactions": [
    {
      "L1Info": {
        "amountSuccess": true,
        "depositAmount": "1000000000000000000",
        "depositAmountSuccess": true,
        "ethereumBlockNum": 8291779,
        "historicDepositAmountUSD": 1688.25,
        "toForgeL1TransactionsNum": 624,
        "userOrigin": true
      },
      "L1orL2": "L1",
      "L2Info": null,
      "amount": "0",
      "batchNum": 625,
      "fromAccountIndex": "hez:ETH:1419",
      "fromBJJ": "hez:jOfSX1JbZ-uixzMYL1jWb3ghoioGbEvayijeMeSekYy_",
      "fromHezEthereumAddress": "hez:0x29fd05074EEB02df660E5f2247cc80DE43E70Aff",
      "historicUSD": null,
      "id": "0x00916f52a44556a7e66de7b26688797535002205b911744440e69ccae7e98b1503",
      "itemId": 6640,
      "position": 0,
      "timestamp": "2021-03-24T18:21:23Z",
      "toAccountIndex": "hez:ETH:0",
      "toBJJ": null,
      "toHezEthereumAddress": null,
      "token": {
        "USD": 1566.7,
        "decimals": 18,
        "ethereumAddress": "0x0000000000000000000000000000000000000000",
        "ethereumBlockNum": 0,
        "fiatUpdate": "2021-03-23T10:55:20.984541Z",
        "id": 0,
        "itemId": 1,
        "name": "Ether",
        "symbol": "ETH"
      },
      "type": "CreateAccountDeposit"
    },
    {
      "L1Info": {
        "amountSuccess": true,
        "depositAmount": "1959275309000000000",
        "depositAmountSuccess": true,
        "ethereumBlockNum": 8291781,
        "historicDepositAmountUSD": 13.585614992606,
        "toForgeL1TransactionsNum": 624,
        "userOrigin": true
      },
      "L1orL2": "L1",
      "L2Info": null,
      "amount": "0",
      "batchNum": 625,
      "fromAccountIndex": "hez:HEZ:1420",
      "fromBJJ": "hez:jOfSX1JbZ-uixzMYL1jWb3ghoioGbEvayijeMeSekYy_",
      "fromHezEthereumAddress": "hez:0x29fd05074EEB02df660E5f2247cc80DE43E70Aff",
      "historicUSD": null,
      "id": "0x00821a392933240f5f468b6e44e0d8d405cf78e380a4d2699d01d7c28d257042e6",
      "itemId": 6641,
      "position": 1,
      "timestamp": "2021-03-24T18:21:53Z",
      "toAccountIndex": "hez:HEZ:0",
      "toBJJ": null,
      "toHezEthereumAddress": null,
      "token": {
        "USD": 5.9725,
        "decimals": 18,
        "ethereumAddress": "0x2521bc90b4f5fb9a8d61278197e5ff5cdbc4fbf2",
        "ethereumBlockNum": 8255859,
        "fiatUpdate": "2021-03-23T10:55:20.939031Z",
        "id": 1,
        "itemId": 2,
        "name": "Hermez Network Token",
        "symbol": "HEZ"
      },
      "type": "CreateAccountDeposit"
    },
    {
      "L1Info": {
        "amountSuccess": true,
        "depositAmount": "1000000000000000000",
        "depositAmountSuccess": true,
        "ethereumBlockNum": 8291784,
        "historicDepositAmountUSD": 1688.25,
        "toForgeL1TransactionsNum": 624,
        "userOrigin": true
      },
      "L1orL2": "L1",
      "L2Info": null,
      "amount": "0",
      "batchNum": 625,
      "fromAccountIndex": "hez:ETH:1421",
      "fromBJJ": "hez:2kuiruc1QJdhR37zvFNnHFmzOOWeA4VgpbkxjBzeGaGR",
      "fromHezEthereumAddress": "hez:0x73c0E1f921dDCcfD7Baf634ff7Df40314C8BfbF7",
      "historicUSD": null,
      "id": "0x001cf9099cc90012f7539a6a490086b4efeea27991976bc737d89a1aad2757f8fa",
      "itemId": 6642,
      "position": 2,
      "timestamp": "2021-03-24T18:22:38Z",
      "toAccountIndex": "hez:ETH:0",
      "toBJJ": null,
      "toHezEthereumAddress": null,
      "token": {
        "USD": 1566.7,
        "decimals": 18,
        "ethereumAddress": "0x0000000000000000000000000000000000000000",
        "ethereumBlockNum": 0,
        "fiatUpdate": "2021-03-23T10:55:20.984541Z",
        "id": 0,
        "itemId": 1,
        "name": "Ether",
        "symbol": "ETH"
      },
      "type": "CreateAccountDeposit"
    },
    {
      "L1Info": {
        "amountSuccess": true,
        "depositAmount": "1956290403000000000",
        "depositAmountSuccess": true,
        "ethereumBlockNum": 8291785,
        "historicDepositAmountUSD": 13.564917654402,
        "toForgeL1TransactionsNum": 624,
        "userOrigin": true
      },
      "L1orL2": "L1",
      "L2Info": null,
      "amount": "0",
      "batchNum": 625,
      "fromAccountIndex": "hez:HEZ:1422",
      "fromBJJ": "hez:2kuiruc1QJdhR37zvFNnHFmzOOWeA4VgpbkxjBzeGaGR",
      "fromHezEthereumAddress": "hez:0x73c0E1f921dDCcfD7Baf634ff7Df40314C8BfbF7",
      "historicUSD": null,
      "id": "0x00acc7dac7929db647e031495160bc4dc79378e6c836212bdedfa95401655f49ea",
      "itemId": 6643,
      "position": 3,
      "timestamp": "2021-03-24T18:22:53Z",
      "toAccountIndex": "hez:HEZ:0",
      "toBJJ": null,
      "toHezEthereumAddress": null,
      "token": {
        "USD": 5.9725,
        "decimals": 18,
        "ethereumAddress": "0x2521bc90b4f5fb9a8d61278197e5ff5cdbc4fbf2",
        "ethereumBlockNum": 8255859,
        "fiatUpdate": "2021-03-23T10:55:20.939031Z",
        "id": 1,
        "itemId": 2,
        "name": "Hermez Network Token",
        "symbol": "HEZ"
      },
      "type": "CreateAccountDeposit"
    },
    {
      "L1Info": {
        "amountSuccess": true,
        "depositAmount": "1000000000000000000",
        "depositAmountSuccess": true,
        "ethereumBlockNum": 8291792,
        "historicDepositAmountUSD": 1686.79,
        "toForgeL1TransactionsNum": 624,
        "userOrigin": true
      },
      "L1orL2": "L1",
      "L2Info": null,
      "amount": "0",
      "batchNum": 625,
      "fromAccountIndex": "hez:ETH:1423",
      "fromBJJ": "hez:GTvXaT_eQDDMr9u70G1GEUwd_aSFCnxYgURGhRb6ix58",
      "fromHezEthereumAddress": "hez:0x0B9270Ef6D4487B50FB364005C939121fC5C3f21",
      "historicUSD": null,
      "id": "0x00185c403912ce2e888b7fe99a3ad169db90af86c96b80911a5e16b49a484416ac",
      "itemId": 6644,
      "position": 4,
      "timestamp": "2021-03-24T18:24:38Z",
      "toAccountIndex": "hez:ETH:0",
      "toBJJ": null,
      "toHezEthereumAddress": null,
      "token": {
        "USD": 1566.7,
        "decimals": 18,
        "ethereumAddress": "0x0000000000000000000000000000000000000000",
        "ethereumBlockNum": 0,
        "fiatUpdate": "2021-03-23T10:55:20.984541Z",
        "id": 0,
        "itemId": 1,
        "name": "Ether",
        "symbol": "ETH"
      },
      "type": "CreateAccountDeposit"
    },
    {
      "L1Info": {
        "amountSuccess": true,
        "depositAmount": "2100000000000000000",
        "depositAmountSuccess": true,
        "ethereumBlockNum": 8291793,
        "historicDepositAmountUSD": 14.5614,
        "toForgeL1TransactionsNum": 624,
        "userOrigin": true
      },
      "L1orL2": "L1",
      "L2Info": null,
      "amount": "0",
      "batchNum": 625,
      "fromAccountIndex": "hez:HEZ:1424",
      "fromBJJ": "hez:GTvXaT_eQDDMr9u70G1GEUwd_aSFCnxYgURGhRb6ix58",
      "fromHezEthereumAddress": "hez:0x0B9270Ef6D4487B50FB364005C939121fC5C3f21",
      "historicUSD": null,
      "id": "0x00e77df43052fddee4125c1691ba00c5d86b32a23e54edae1e1ef6dabf2f7d1ff6",
      "itemId": 6645,
      "position": 5,
      "timestamp": "2021-03-24T18:24:53Z",
      "toAccountIndex": "hez:HEZ:0",
      "toBJJ": null,
      "toHezEthereumAddress": null,
      "token": {
        "USD": 5.9725,
        "decimals": 18,
        "ethereumAddress": "0x2521bc90b4f5fb9a8d61278197e5ff5cdbc4fbf2",
        "ethereumBlockNum": 8255859,
        "fiatUpdate": "2021-03-23T10:55:20.939031Z",
        "id": 1,
        "itemId": 2,
        "name": "Hermez Network Token",
        "symbol": "HEZ"
      },
      "type": "CreateAccountDeposit"
    },
    {
      "L1Info": null,
      "L1orL2": "L2",
      "L2Info": {
        "fee": 32,
        "historicFeeUSD": 1.316578125,
        "nonce": 0
      },
      "amount": "200000000000000000",
      "batchNum": 625,
      "fromAccountIndex": "hez:ETH:1418",
      "fromBJJ": "hez:TMPdFi1QuJTTVFytspacYQl1Va5e8WCkSk4qUaq65JEA",
      "fromHezEthereumAddress": "hez:0x91b7377A3Dd7931Ad4757B60d7E859954B939E85",
      "historicUSD": 337.044,
      "id": "0x02fb81fd640d833661369586823fb66651d937765b7fb315c6b4069db2cbeb7c76",
      "itemId": 6646,
      "position": 6,
      "timestamp": "2021-03-24T18:38:23Z",
      "toAccountIndex": "hez:ETH:1418",
      "toBJJ": "hez:TMPdFi1QuJTTVFytspacYQl1Va5e8WCkSk4qUaq65JEA",
      "toHezEthereumAddress": "hez:0x91b7377A3Dd7931Ad4757B60d7E859954B939E85",
      "token": {
        "USD": 1566.7,
        "decimals": 18,
        "ethereumAddress": "0x0000000000000000000000000000000000000000",
        "ethereumBlockNum": 0,
        "fiatUpdate": "2021-03-23T10:55:20.984541Z",
        "id": 0,
        "itemId": 1,
        "name": "Ether",
        "symbol": "ETH"
      },
      "type": "Transfer"
    }
  ],
  "pendingItems": 0
}
`
	hash = "0x0237b03cb0bf99f27cb3080dd734896f350057bff36c4ae98b6f0dc6fd86f2d9ad"
)

func Test_run(t *testing.T) {
	nodeURL := "https://hermez.io"
	chainID := uint16(5)

	gock.New(nodeURL).
		Get("v1/transactions-history").
		Reply(200).
		BodyString(txHistory)

	gock.New(nodeURL).
		Get("v1/batches").
		Reply(200).
		BodyString(lastBatch)

	gock.New(nodeURL).
		Get("v1/accounts").
		Reply(200).
		BodyString(account)

	gock.New(nodeURL).
		Post("v1/transactions-pool").
		Reply(200).
		BodyString(hash)

	if err := run(nodeURL, chainID); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}
