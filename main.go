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

	//mnemonic := "seat mandate concert notable miss worth bottom inquiry find raven seat pilot office foam unique"
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
	amount := big.NewInt(7000000000000000)
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

	// Create a transfer to baby jubjub address
	toBJJAddr := "hez:rkv1d1K9P9sNW9AxbndYL7Ttgtqros4Rwgtw9ewJ-S_b"
	err = transferToBjj(bjj, c, chainID, toBJJAddr, amount, fee)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	// Create a transfer to ethereum address
	toEthAddr := "0xd9391B20559777E1b94954Ed84c28541E35bFEb8"
	err = transferToEthAddress(bjj, c, chainID, toEthAddr, amount, fee)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	// Create a transfer to idx address
	toIdx := hezCommon.Idx(1276)
	err = transfer(bjj, c, chainID, toIdx, amount, fee)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	// Create a exit transfer
	return exit(bjj, c, chainID, amount, fee)
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
	return nil
}

// transferToBjj create and send a transfer to baby jubjub transaction
func transferToBjj(bjj *hermez.Wallet, c *client.Client, chainID uint16, toBjjAddr string,
	amount *big.Int, fee hezCommon.FeeSelector) error {
	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, bjj.HezBjjAddress, amount, token.TokenID)

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
		return err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return err
	}
	logger.Info("Tx Sent", logger.Params{"hash": hash})
	return nil
}

// getAccountInfo fetches account from network, check the balance and returns
// the idx, nonce and an error if occurs
func getAccountInfo(c *client.Client, HezBjjAddress string, amount *big.Int,
	tokenID hezCommon.TokenID) (hezCommon.Idx, hezCommon.Nonce, error) {

	idx := hezCommon.Idx(0)
	nonce := hezCommon.Nonce(0)

	// Get account to fetch the user idx, nonce and balance
	ac, err := c.GetAccount(HezBjjAddress, tokenID)
	if err != nil {
		return idx, nonce, err
	}
	ethAc, err := ac.Accounts.GetAccount(tokenID)
	if err != nil {
		return idx, nonce, err
	}
	logger.Info("Account", logger.Params{"address": HezBjjAddress, "address_idx": ethAc.Idx})

	idx = hezCommon.Idx(ethAc.Idx)
	nonce = ethAc.Nonce

	// Create the transaction
	balance := ac.Accounts[0].Balance
	if balance.Cmp(amount) < 0 {
		return idx, nonce, errors.E("invalid amount", errors.Params{"balance": balance, "amount": amount})
	}
	return idx, nonce, nil
}

// transferToEthAddress create and send a transfer to ethereum address transaction
func transferToEthAddress(bjj *hermez.Wallet, c *client.Client, chainID uint16,
	toEthAddr string, amount *big.Int, fee hezCommon.FeeSelector) error {

	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, bjj.HezBjjAddress, amount, token.TokenID)

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
		return err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return err
	}
	logger.Info("Tx Sent", logger.Params{"hash": hash})
	return nil
}

// transfer create and send a transfer transaction
func transfer(bjj *hermez.Wallet, c *client.Client, chainID uint16, toIdx hezCommon.Idx,
	amount *big.Int, fee hezCommon.FeeSelector) error {

	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, bjj.HezBjjAddress, amount, token.TokenID)

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
		return err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return err
	}
	logger.Info("Tx Sent", logger.Params{"hash": hash})
	return nil
}

// exit create and send a transfer exit transaction
func exit(bjj *hermez.Wallet, c *client.Client, chainID uint16, amount *big.Int,
	fee hezCommon.FeeSelector) error {

	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, bjj.HezBjjAddress, amount, token.TokenID)

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
		return err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return err
	}
	logger.Info("Tx Sent", logger.Params{"hash": hash})
	return nil
}
