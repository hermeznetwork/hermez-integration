package main

import (
	"encoding/hex"
	"math/big"
	"time"

	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
	"github.com/hermeznetwork/hermez-integration/hermez"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

const (
	// Rinkeby chain id
	chainID = uint16(4)
	nodeURL = "https://api.testnet.hermez.io"
)

func main() {
	err := run(nodeURL, chainID)
	if err != nil {
		logger.Fatal(err)
	}
}

func run(nodeURL string, chainID uint16) error {
	// create a new Hermez node client
	c := client.New(nodeURL)

	// get supported tokens
	tokens, err := c.GetTokens()
	if err != nil {
		return err
	}
	for _, t := range tokens.Tokens {
		logger.Info("Token "+t.Name, logger.Params{
			"TokenID":     t.TokenID,
			"Decimals":    t.Decimals,
			"EthAddr":     t.EthAddr,
			"EthBlockNum": t.EthBlockNum,
			"ItemID":      t.ItemID,
			"Symbol":      t.Symbol,
			"USD":         t.USD,
			"USDUpdate":   t.USDUpdate,
		})
	}

	// track incoming deposits
	err = deposits(c)
	if err != nil {
		return err
	}

	mnemonic := "butter embrace sunny tilt soap where soul finish shop west rough flock"

	// Increase the wallet index to generate a new wallet based
	// in the bip39, starting from zero
	walletIndex := 0

	// Create a baby jujub wallet based in the mnemonic and index
	// After create the wallet, the accounts must generate into the network
	// one for each token for the same wallet, calling the smart contract methods:
	// - CreateAccountDeposit: creates a new token account for wallet
	// - CreateAccountDepositTransfer: creates a new token account for wallet and transfer
	// - TransferToBjj: Transfer to Bjj account, this transaction
	//	encourages the coordinator to create new accounts through the L1 coordinator
	//	transaction CreateAccountBjj.
	//
	// After creating the wallet we must create an account for each token, and we must get
	// the id (IDX) and nonce for this account to create a transfer.
	bjj, err := hermez.NewBJJ(mnemonic, walletIndex)
	if err != nil {
		return err
	}
	pkBuf := [hermez.PkLength]byte(bjj.PrivateKey)
	logger.Info("BJJ Create", logger.Params{
		"hez_eth_address": bjj.HezEthAddress,
		"hez_bjj_address": bjj.HezBjjAddress,
		"private_key":     "0x" + hex.EncodeToString(pkBuf[:]),
	})

	// A fee is a percentage value from the token amount, and the fee amount in USD must
	// be greater than the minimum fee value the coordinator accepts. The fee value in the
	// L2 transaction apply a factor encoded by an index from the transaction fee table:
	// https://docs.hermez.io/#/developers/protocol/hermez-protocol/fee-table?id=transaction-fee-table
	amount := big.NewInt(6000000000000000)
	fee := hezCommon.FeeSelector(126) // 10.2%
	feeAmount, err := hezCommon.CalcFeeAmount(amount, fee)
	if err != nil {
		return err
	}

	logger.Info("Fee", logger.Params{
		"amount_wei":     amount.String(),
		"amount_eth":     hermez.WeiToEther(amount).String(),
		"fee_selector":   fee,
		"fee_percentage": fee.Percentage(),
		"fee_amount_wei": feeAmount.String(),
		"fee_amount_eth": hermez.WeiToEther(feeAmount).String(),
	})

	// Create a transfer to a idx account
	toIdx := hezCommon.Idx(1276)
	txID, err := transfer(bjj, c, chainID, toIdx, amount, fee)
	if err != nil {
		return err
	}
	logger.Info("transferToBjj", logger.Params{"tx_id": txID})

	poolTx, err := c.GetPoolTx(txID)
	if err != nil {
		return err
	}
	logger.Info("GetPoolTx", logger.Params{"tx_id": poolTx.TxID, "state": poolTx.State})

	// Create a transfer to baby jubjub address
	time.Sleep(5 * time.Second)
	toBJJAddr := "hez:rkv1d1K9P9sNW9AxbndYL7Ttgtqros4Rwgtw9ewJ-S_b"
	txID, err = transferToBjj(bjj, c, chainID, toBJJAddr, amount, fee)
	if err != nil {
		return err
	}
	logger.Info("transferToBjj", logger.Params{"tx_id": txID})

	// Create a transfer to ethereum address
	time.Sleep(5 * time.Second)
	toEthAddr := "0xd9391B20559777E1b94954Ed84c28541E35bFEb8"
	txID, err = transferToEthAddress(bjj, c, chainID, toEthAddr, amount, fee)
	if err != nil {
		return err
	}
	logger.Info("transferToEthAddress", logger.Params{"tx_id": txID})

	// Create a exit transfer
	time.Sleep(5 * time.Second)
	txID, err = exit(bjj, c, chainID, amount, fee)
	if err != nil {
		return err
	}
	logger.Info("exit", logger.Params{"tx_id": txID})

	return nil
}

// deposits get last batch number and get the transactions
func deposits(c *client.Client) error {
	// Fetch the last batch for we can pulling the transactions
	lastBatch, err := c.GetLastBatch()
	if err != nil {
		return err
	}
	logger.Info("Last Batch", logger.Params{"last_batch": lastBatch.BatchNum})

	// Get all transactions for a batch for tracking the deposits
	batch, err := c.GetBatchTxs(lastBatch.BatchNum)
	if err != nil {
		return err
	}
	logger.Info("Batch", logger.Params{"batch": lastBatch.BatchNum, "txs": len(batch.Txs)})

	// Get an specific transaction by id
	txID := "0x021f6c6fa724d14f784cf536b0bd46388b70ead9281a19bebbb8688effe16a8b65"
	tx, err := c.GetTx(txID)
	if err != nil {
		return err
	}
	logger.Info("GetTx", logger.Params{"tx_id": tx.TxID, "batch_number": tx.BatchNum})

	return nil
}

// getAccountInfo fetches account from network, check the balance and returns
// the idx, nonce and an error if occurs
func getAccountInfo(c *client.Client, bjjAddress, hezEthAddress *string, amount *big.Int,
	tokenID hezCommon.TokenID) (hezCommon.Idx, hezCommon.Nonce, error) {

	idx := hezCommon.Idx(0)
	nonce := hezCommon.Nonce(0)

	// Get account to fetch the user idx, nonce and balance
	ac, err := c.GetAccount(bjjAddress, hezEthAddress, tokenID)
	if err != nil {
		return idx, nonce, err
	}
	ethAc, err := ac.Accounts.GetAccount(tokenID)
	if err != nil {
		return idx, nonce, err
	}
	logger.Info("Account", logger.Params{"bjjAddress": bjjAddress,
		"hezEthAddress": hezEthAddress, "address_idx": ethAc.Idx})

	idx = hezCommon.Idx(ethAc.Idx)
	nonce = ethAc.Nonce

	// Create the transaction
	balance := ac.Accounts[0].Balance
	if balance.Cmp(amount) < 0 {
		return idx, nonce, errors.E("invalid amount", errors.Params{"balance": balance, "amount": amount})
	}
	return idx, nonce, nil
}

// transfer create and send a transfer transaction
func transfer(bjj *hermez.Wallet, c *client.Client, chainID uint16, toIdx hezCommon.Idx,
	amount *big.Int, fee hezCommon.FeeSelector) (string, error) {

	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)

	tx, err := hermez.CreateTransfer(
		chainID,
		toIdx,
		amount,
		bjj.PrivateKey,
		idx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return "", err
	}
	logger.Info("Tx Sent", logger.Params{"tx_id": hash})

	return hash, nil
}

// transferToBjj create and send a transfer to baby jubjub transaction
func transferToBjj(bjj *hermez.Wallet, c *client.Client, chainID uint16, toBjjAddr string,
	amount *big.Int, fee hezCommon.FeeSelector) (string, error) {
	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)

	tx, err := hermez.CreateTransferToBjj(
		chainID,
		toBjjAddr,
		amount,
		bjj.PrivateKey,
		idx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return "", err
	}
	logger.Info("Tx Sent", logger.Params{"tx_id": hash})

	return hash, nil
}

// transferToEthAddress create and send a transfer to ethereum address transaction
func transferToEthAddress(bjj *hermez.Wallet, c *client.Client, chainID uint16,
	toEthAddr string, amount *big.Int, fee hezCommon.FeeSelector) (string, error) {

	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)

	tx, err := hermez.CreateTransferToEthAddress(
		chainID,
		toEthAddr,
		amount,
		bjj.PrivateKey,
		idx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return "", err
	}
	logger.Info("Tx Sent", logger.Params{"tx_id": hash})

	return hash, nil
}

// exit create and send a transfer exit transaction
func exit(bjj *hermez.Wallet, c *client.Client, chainID uint16, amount *big.Int,
	fee hezCommon.FeeSelector) (string, error) {

	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)

	tx, err := hermez.CreateExit(
		chainID,
		amount,
		bjj.PrivateKey,
		idx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return "", err
	}
	logger.Info("Tx Sent", logger.Params{"tx_id": hash})

	return hash, nil
}
