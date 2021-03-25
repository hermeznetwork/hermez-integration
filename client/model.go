package client

import (
	"encoding/json"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/apitypes"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type (
	// Account is a representation of a account with additional information
	// required by the API
	Account struct {
		ItemID           uint64              `meddler:"item_id"`
		Idx              hezCommon.Idx       `meddler:"idx"`
		BatchNum         hezCommon.BatchNum  `meddler:"batch_num"`
		PublicKey        apitypes.HezBJJ     `meddler:"bjj"`
		EthAddr          apitypes.HezEthAddr `meddler:"eth_addr"`
		Nonce            hezCommon.Nonce     `meddler:"nonce"`   // max of 40 bits used
		Balance          *BigInt             `meddler:"balance"` // max of 192 bits used
		TotalItems       uint64              `meddler:"total_items"`
		FirstItem        uint64              `meddler:"first_item"`
		LastItem         uint64              `meddler:"last_item"`
		TokenID          hezCommon.TokenID   `meddler:"token_id"`
		TokenItemID      int                 `meddler:"token_item_id"`
		TokenEthBlockNum int64               `meddler:"token_block"`
		TokenEthAddr     ethCommon.Address   `meddler:"token_eth_addr"`
		TokenName        string              `meddler:"name"`
		TokenSymbol      string              `meddler:"symbol"`
		TokenDecimals    uint64              `meddler:"decimals"`
		TokenUSD         *float64            `meddler:"usd"`
		TokenUSDUpdate   *time.Time          `meddler:"usd_update"`
	}
	// Accounts is a representation of a account API response.
	Accounts struct {
		Accounts     []Account `json:"accounts"`
		PendingItems uint64    `json:"pendingItems"`
	}
	// Txs is a representation of a tx history API response.
	Txs struct {
		Txs          []historydb.TxAPI `json:"transactions"`
		PendingItems uint64            `json:"pendingItems"`
	}
	// Batches is a representation of a batches API response.
	Batches struct {
		Batches      []historydb.BatchAPI `json:"batches"`
		PendingItems uint64               `json:"pendingItems"`
	}
	// Tx is a representation of a transaction API request.
	Tx struct {
		TxID      hezCommon.TxID        `json:"id"`
		Type      hezCommon.TxType      `json:"type"`
		TokenID   hezCommon.TokenID     `json:"tokenId"`
		FromIdx   string                `json:"fromAccountIndex"`
		ToIdx     string                `json:"toAccountIndex"`
		ToEthAddr string                `json:"toHezEthereumAddress"`
		ToBJJ     string                `json:"toBjj"`
		Amount    string                `json:"amount"`
		Fee       hezCommon.FeeSelector `json:"fee"`
		Nonce     hezCommon.Nonce       `json:"nonce"`
		Signature babyjub.SignatureComp `json:"signature"`
	}
	// BigInt is big.Int wrapper
	BigInt struct {
		big.Int
	}
)

// UnmarshalJSON unmarshal BitInt object
func (i *BigInt) UnmarshalJSON(b []byte) error {
	var val string
	err := json.Unmarshal(b, &val)
	if err != nil {
		return err
	}

	i.SetString(val, 10)
	return nil
}
