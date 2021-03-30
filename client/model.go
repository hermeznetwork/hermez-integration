package client

import (
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/Pantani/errors"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

type (
	// Batch is a representation of a batch with additional information
	// required by the API, and extracted by joining block table
	Batch struct {
		ItemID        uint64             `json:"itemId"`
		BatchNum      hezCommon.BatchNum `json:"batchNum"`
		EthBlockNum   int64              `json:"ethereumBlockNum"`
		EthBlockHash  ethCommon.Hash     `json:"ethereumBlockHash"`
		Timestamp     time.Time          `json:"timestamp"`
		ForgerAddr    ethCommon.Address  `json:"forgerAddr"`
		CollectedFees CollectedFees      `json:"collectedFees"`
		TotalFeesUSD  *float64           `json:"historicTotalCollectedFeesUSD"`
		StateRoot     apitypes.BigIntStr `json:"stateRoot"`
		NumAccounts   int                `json:"numAccounts"`
		ExitRoot      apitypes.BigIntStr `json:"exitRoot"`
		ForgeL1TxsNum *int64             `json:"forgeL1TransactionsNum"`
		SlotNum       int64              `json:"slotNum"`
		ForgedTxs     int                `json:"forgedTransactions"`
	}

	// AccountAPI is a representation of a account API response.
	AccountAPI struct {
		Accounts     Accounts `json:"accounts"`
		PendingItems uint64   `json:"pendingItems"`
	}

	// TxAPI is a representation of a tx history API response.
	TxAPI struct {
		Txs          []historydb.TxAPI `json:"transactions"`
		PendingItems uint64            `json:"pendingItems"`
	}

	// BatchAPI is a representation of a batches API response.
	BatchAPI struct {
		Batches      []Batch `json:"batches"`
		PendingItems uint64  `json:"pendingItems"`
	}

	// TokenAPI is a representation of a tokens API response.
	TokenAPI struct {
		Tokens       []Token `json:"tokens"`
		PendingItems uint64  `json:"pendingItems"`
	}

	// Token is a representation of a tokens API object.
	Token struct {
		ItemID      uint64            `json:"itemId"`
		TokenID     hezCommon.TokenID `json:"id"`
		EthBlockNum int64             `json:"ethereumBlockNum"`
		EthAddr     ethCommon.Address `json:"ethereumAddress"`
		Name        string            `json:"name"`
		Symbol      string            `json:"symbol"`
		Decimals    uint64            `json:"decimals"`
		USD         float64           `json:"USD"`
		USDUpdate   time.Time         `json:"fiatUpdate"`
	}

	// Account is a representation of a account with additional information
	// required by the API
	Account struct {
		ItemID    uint64              `json:"itemId"`
		Idx       StrHezIdx           `json:"accountIndex"`
		BatchNum  hezCommon.BatchNum  `json:"batch_num"`
		PublicKey apitypes.HezBJJ     `json:"bjj"`
		EthAddr   apitypes.HezEthAddr `json:"hezEthereumAddress"`
		Nonce     hezCommon.Nonce     `json:"nonce"`
		Balance   *BigInt             `json:"balance"`
		Token     hezCommon.Token     `json:"token"`
	}

	// Tx is a representation of a transaction API request.
	Tx struct {
		TxID      hezCommon.TxID `json:"id" binding:"required"`
		Type      string         `json:"type"`
		TokenID   uint32         `json:"tokenId"`
		FromIdx   string         `json:"fromAccountIndex" binding:"required"`
		ToIdx     string         `json:"toAccountIndex"`
		ToEthAddr string         `json:"toHezEthereumAddress"`
		ToBJJ     string         `json:"toBjj"`
		Amount    string         `json:"amount" binding:"required"`
		Fee       uint64         `json:"fee"`
		Nonce     uint64         `json:"nonce"`
		Signature string         `json:"signature"`
	}

	// Accounts is a representation of a account list.
	Accounts []Account

	// CollectedFees is used to retrieve common.batch.CollectedFee from the DB
	CollectedFees map[hezCommon.TokenID]BigInt

	// BigInt is big.Int wrapper
	BigInt struct {
		big.Int
	}

	// StrHezIdx is used to unmarshal HezIdx directly into an alias of common.Idx
	StrHezIdx hezCommon.Idx
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

// UnmarshalText unmarshal a StrHezIdx
func (s *StrHezIdx) UnmarshalText(text []byte) error {
	withoutHez := strings.TrimPrefix(string(text), "hez:")
	splitted := strings.Split(withoutHez, ":")
	const expectedLen = 2
	if len(splitted) != expectedLen {
		return errors.E("cannot unmarshal into StrHezIdx", errors.Params{"text": text})
	}
	idxInt, err := strconv.Atoi(splitted[1])
	if err != nil {
		return err
	}
	*s = StrHezIdx(hezCommon.Idx(idxInt))
	return nil
}

// GetAccount get account by token ID
func (acs *Accounts) GetAccount(tokenID hezCommon.TokenID) (Account, error) {
	for _, ac := range *acs {
		if ac.Token.TokenID == tokenID {
			return ac, nil
		}
	}
	return Account{}, errors.E("account not found",
		errors.Params{"token_id": tokenID})
}
