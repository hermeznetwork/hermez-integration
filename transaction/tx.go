package transaction

import (
	"math/big"

	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
	"github.com/hermeznetwork/hermez-integration/hermez"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

// GetAccountInfo fetches account from network, check the balance and returns
// the idx, nonce and an error if occurs
func GetAccountInfo(c *client.Client, bjjAddress, hezEthAddress *string, amount *big.Int,
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

// Transfer create and send a Transfer transaction
func Transfer(bjj *hermez.Wallet, c *client.Client, chainID uint16, toIdx hezCommon.Idx,
	amount *big.Int, fee hezCommon.FeeSelector, token hezCommon.Token) (string, error) {

	// Get account idx, nonce and check the balance
	idx, nonce, err := GetAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)
	if err != nil {
		return "", err
	}

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
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	logger.Info("Tx Sent", logger.Params{"tx_id": hash})

	return hash, nil
}

// TransferToBjj create and send a Transfer to baby jubjub transaction
func TransferToBjj(bjj *hermez.Wallet, c *client.Client, chainID uint16, toBjjAddr string,
	amount *big.Int, fee hezCommon.FeeSelector, token hezCommon.Token) (string, error) {
	// Get account idx, nonce and check the balance
	idx, nonce, err := GetAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)
	if err != nil {
		return "", err
	}

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
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// TransferToEthAddress create and send a Transfer to ethereum address transaction
func TransferToEthAddress(bjj *hermez.Wallet, c *client.Client, chainID uint16,
	toEthAddr string, amount *big.Int, fee hezCommon.FeeSelector, token hezCommon.Token) (string, error) {

	// Get account idx, nonce and check the balance
	idx, nonce, err := GetAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)
	if err != nil {
		return "", err
	}

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
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// Exit create and send a Transfer Exit transaction
func Exit(bjj *hermez.Wallet, c *client.Client, chainID uint16, amount *big.Int,
	fee hezCommon.FeeSelector, token hezCommon.Token) (string, error) {

	// Get account idx, nonce and check the balance
	idx, nonce, err := GetAccountInfo(c, &bjj.HezBjjAddress, nil, amount, token.TokenID)
	if err != nil {
		return "", err
	}

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
	hash, err := c.SendTransaction(*tx, token)
	if err != nil {
		return "", err
	}
	return hash, nil
}
