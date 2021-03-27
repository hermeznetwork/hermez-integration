package main

import (
	"encoding/hex"
	"math/big"

	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
	"github.com/hermeznetwork/hermez-integration/hermez"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

func main() {
	// Rinkeby chain id
	chainID := uint16(4)
	err := run("http://localhost:8086", chainID)
	if err != nil {
		logger.Fatal(err)
	}
}

func run(nodeURL string, chainID uint16) error {
	// wallet configs
	mnemonic := "seat mandate concert notable miss worth bottom inquiry find raven seat pilot office foam unique"
	toBjjAddr := "hez:rkv1d1K9P9sNW9AxbndYL7Ttgtqros4Rwgtw9ewJ-S_b"
	amount := big.NewInt(100)

	// Increase the wallet index to generate a new wallet based
	// in the bip39, starting from zero
	walletIndex := 0

	// Create a wallet
	bjj, err := hermez.CreateBJJFromEthAddr(mnemonic, walletIndex)
	if err != nil {
		return err
	}

	// Create the Baby Jubjub hez address
	compressPk := bjj.Public().Compress().String()
	b, err := hex.DecodeString(compressPk)
	if err != nil {
		return err
	}
	address := hermez.NewHezBJJ(bjj.Public().Compress())
	pkBuf := [hermez.PkLength]byte(bjj)

	// Create the Hermez hez address
	pkBytes := hezCommon.SwapEndianness(b)
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkBytes[:])

	logger.Info("BJJ Create", logger.Params{
		"address_encoded": address,
		"address_hex":     "0x" + pkComp.String(),
		"private key":     "0x" + hex.EncodeToString(pkBuf[:]),
	})

	// The accounts must generate one for each token for the same wallet, calling the
	// smart contract methods:
	// - CreateAccountDeposit: creates a new token account for wallet
	// - CreateAccountDepositTransfer: creates a new token account for wallet and transfer
	// - TransferToBjj: Transfer to Bjj account, this transaction
	//	encourages the coordinator to create new accounts through the L1 coordinator
	//	transaction CreateAccountBjj.
	//
	// After creating the wallet we must create an account for each token, and we must get
	// the id (IDX) and nonce for this account to create a transfer.
	c := client.New(nodeURL)
	ac, err := c.GetAccount(address, hermez.EthToken.TokenID)
	if err != nil {
		return err
	}
	ethAc, err := ac.Accounts.GetAccount(hermez.EthToken.TokenID)
	if err != nil {
		return err
	}
	logger.Info("Account", logger.Params{"address": address, "address_idx": ethAc.Idx})

	// Create the transaction (TransferToBjj)
	balance := ac.Accounts[0].Balance
	if balance.Cmp(amount) < 0 {
		return errors.E("invalid amount", errors.Params{"balance": balance, "amount": amount})
	}

	idx := hezCommon.Idx(ethAc.Idx)
	tx, err := hermez.CreateTx(
		chainID,
		toBjjAddr,
		amount,
		bjj,
		idx,
		ethAc.Token.TokenID,
		ethAc.Nonce,
		235,
	)
	if err != nil {
		return err
	}
	logger.Info("Tx Created", logger.Params{"hash": tx.TxID, "tx": tx})

	// Send the transaction
	hash, err := c.SendTransaction(*tx, hermez.EthToken)
	if err != nil {
		return err
	}
	logger.Info("Tx Sent", logger.Params{"hash": hash})

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
