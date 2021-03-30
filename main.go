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

	// track incoming deposits
	err := deposits(c)
	if err != nil {
		return err
	}

	mnemonic := "seat mandate concert notable miss worth bottom inquiry find raven seat pilot office foam unique"

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
		"address_encoded": bjj.HezAddress,
		"address_hex":     "0x" + bjj.PublicKey.String(),
		"private key":     "0x" + hex.EncodeToString(pkBuf[:]),
	})

	// Create a transfer to baby jubjub address
	toBJJAddr := "hez:rkv1d1K9P9sNW9AxbndYL7Ttgtqros4Rwgtw9ewJ-S_b"
	amount := big.NewInt(1000)
	err = transferToBjj(bjj, c, chainID, toBJJAddr, amount, 232)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	// Create a transfer to ethereum address
	toEthAddr := "0xd9391B20559777E1b94954Ed84c28541E35bFEb8"
	amount = big.NewInt(1001)
	err = transferToEthAddress(bjj, c, chainID, toEthAddr, amount, 232)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	// Create a transfer to idx address
	toIdx := hezCommon.Idx(1276)
	amount = big.NewInt(1002)
	err = transfer(bjj, c, chainID, toIdx, amount, 232)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	// Create a exit transfer
	amount = big.NewInt(2000)
	return exit(bjj, c, chainID, amount, 231)
}

// deposits get last batch number and get the transactions
func deposits(c *client.Client) error {
	// Fetch the last batch for we can pulling the transactions
	lastBatch, err := c.GetLastBatch()
	if err != nil {
		return err
	}
	logger.Info("Last Batch", logger.Params{"last_batch": lastBatch})

	// Get all transactions for a batch for tracking the deposits
	batch, err := c.GetBatchTxs(lastBatch.BatchNum)
	if err != nil {
		return err
	}
	logger.Info("Batch", logger.Params{"number": lastBatch.BatchNum, "batch": batch})
	return nil
}

// transferToBjj create and send a transfer to baby jubjub transaction
func transferToBjj(bjj *hermez.Wallet, c *client.Client, chainID uint16, toBjjAddr string,
	amount *big.Int, fee hezCommon.FeeSelector) error {
	// Get account idx, nonce and check the balance
	token := hermez.EthToken
	idx, nonce, err := getAccountInfo(c, bjj.HezAddress, amount, token.TokenID)

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
func getAccountInfo(c *client.Client, hezAddress string, amount *big.Int,
	tokenID hezCommon.TokenID) (hezCommon.Idx, hezCommon.Nonce, error) {

	idx := hezCommon.Idx(0)
	nonce := hezCommon.Nonce(0)

	// Get account to fetch the user idx, nonce and balance
	ac, err := c.GetAccount(hezAddress, tokenID)
	if err != nil {
		return idx, nonce, err
	}
	ethAc, err := ac.Accounts.GetAccount(tokenID)
	if err != nil {
		return idx, nonce, err
	}
	logger.Info("Account", logger.Params{"address": hezAddress, "address_idx": ethAc.Idx})

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
	idx, nonce, err := getAccountInfo(c, bjj.HezAddress, amount, token.TokenID)

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
	idx, nonce, err := getAccountInfo(c, bjj.HezAddress, amount, token.TokenID)

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
	idx, nonce, err := getAccountInfo(c, bjj.HezAddress, amount, token.TokenID)

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
