package transaction

import (
	"math/big"
	"strings"

	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
	"github.com/hermeznetwork/hermez-integration/hermez"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

// GetAccountInfo fetches account from network, check the balance and returns
// the idx, nonce and an error if occurs
func GetAccountInfo(c *client.Client, bjjAddress, hezEthAddress *string,
	tokenID hezCommon.TokenID) (hezCommon.Idx, hezCommon.Nonce, error) {

	idx := hezCommon.Idx(0)
	nonce := hezCommon.Nonce(0)

	// Get account to fetch the user idx, nonce and balance
	ac, err := c.GetAccount(bjjAddress, hezEthAddress, tokenID)
	if err != nil {
		return idx, nonce, err
	}
	ethAc, err := ac.Accounts.GetFirstAccount(tokenID)
	if err != nil {
		return idx, nonce, err
	}

	idx = hezCommon.Idx(ethAc.Idx)
	nonce = ethAc.Nonce

	logger.Info("Account selected", logger.Params{"idx": idx, "nonce": nonce})
	return idx, nonce, nil
}

// Transfer create and send a Transfer transaction
func Transfer(bjj *hermez.Wallet, c *client.Client, chainID uint16,
	fromIdx, toIdx hezCommon.Idx, amount *big.Int, fee hezCommon.FeeSelector,
	token hezCommon.Token, nonce hezCommon.Nonce) (string, error) {

	tx, err := hermez.CreateTransfer(
		chainID,
		toIdx,
		amount,
		bjj.PrivateKey,
		fromIdx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// TransferToBjj create and send a Transfer to baby jubjub transaction
func TransferToBjj(bjj *hermez.Wallet, c *client.Client, chainID uint16,
	fromIdx hezCommon.Idx, toBjjAddr string, amount *big.Int, fee hezCommon.FeeSelector,
	token hezCommon.Token, nonce hezCommon.Nonce) (string, error) {

	tx, err := hermez.CreateTransferToBjj(
		chainID,
		toBjjAddr,
		amount,
		bjj.PrivateKey,
		fromIdx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// TransferToEthAddress create and send a Transfer to ethereum address transaction
func TransferToEthAddress(bjj *hermez.Wallet, c *client.Client, chainID uint16,
	fromIdx hezCommon.Idx, toHezEthAddr string, amount *big.Int, fee hezCommon.FeeSelector,
	token hezCommon.Token, nonce hezCommon.Nonce) (string, error) {

	toEthAddr := strings.Replace(toHezEthAddr, "hez:", "", -1)
	tx, err := hermez.CreateTransferToEthAddress(
		chainID,
		toEthAddr,
		amount,
		bjj.PrivateKey,
		fromIdx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Exit create and send a Transfer Exit transaction
func Exit(bjj *hermez.Wallet, c *client.Client, chainID uint16, fromIdx hezCommon.Idx,
	amount *big.Int, fee hezCommon.FeeSelector, token hezCommon.Token,
	nonce hezCommon.Nonce) (string, error) {

	tx, err := hermez.CreateExit(
		chainID,
		amount,
		bjj.PrivateKey,
		fromIdx,
		token.TokenID,
		nonce,
		fee,
	)
	if err != nil {
		return "", err
	}

	// Send the transaction
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}
