package client

import (
	"net/url"
	"strconv"
	"time"

	"github.com/Pantani/errors"
	"github.com/Pantani/request"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

type (
	// Client represents the node API client object
	// https://docs.hermez.io/#/developers/api
	Client struct {
		request request.Request
	}
)

// New creates a new node API client
func New(nodeURL string) *Client {
	return &Client{
		request: request.InitClient(nodeURL),
	}
}

// GetAccount get an account info based in the hermez-integration address and the token id
func (c *Client) GetAccount(bjjAddress string, tokenID hezCommon.TokenID) (*AccountAPI, error) {
	var result *AccountAPI
	err := c.request.GetWithCache(
		&result,
		"v1/accounts",
		url.Values{
			"BJJ":      {bjjAddress},
			"tokenIds": {tokenID.BigInt().String()},
		},
		time.Hour*1,
	)
	if err != nil {
		return nil, err
	}

	if len(result.Accounts) == 0 {
		return nil, errors.E("account not registered",
			errors.Params{"bjj_address": bjjAddress, "token_id": tokenID})
	}
	return result, nil
}

// GetBatchTxs get all transactions history from a batch number
func (c *Client) GetBatchTxs(batchNum hezCommon.BatchNum) (*TxAPI, error) {
	var result *TxAPI
	return result, c.request.GetWithCache(
		&result,
		"v1/transactions-history",
		url.Values{
			"batchNum": {strconv.Itoa(int(batchNum))},
			"order":    {"ASC"},
		},
		time.Hour*1,
	)
}

// GetLastBatch get last Hermez rollup batch
func (c *Client) GetLastBatch() (*BatchAPI, error) {
	var result *Batches
	err := c.request.Get(
		&result,
		"v1/batches",
		url.Values{
			"limit": {"1"},
			"order": {"DESC"},
		},
	)
	if err != nil {
		return nil, err
	}
	if len(result.Batches) == 0 {
		return nil, errors.E("batch not found")
	}

	return &result.Batches[0], nil
}

// SendTransaction send L2 transaction to the coordinator pool
func (c *Client) SendTransaction(tx hezCommon.PoolL2Tx, token hezCommon.Token) (string, error) {
	var result interface{}
	body := NewTxRequest(tx, token)
	err := c.request.Post(&result, "v1/transactions-pool", body)
	if err != nil {
		return "", err
	}
	hash, ok := result.(string)
	if !ok {
		return "", errors.E("invalid tx result",
			errors.Params{"result": result})
	}
	return hash, err
}
