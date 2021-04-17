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
	"github.com/iden3/go-iden3-crypto/babyjub"
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

	// TxHistory is a representation of a transaction history API request.
	TxHistory struct {
		Amount           apitypes.BigIntStr      `json:"amount"`
		Fee              hezCommon.FeeSelector   `json:"fee"`
		FromIdx          StrHezIdx               `json:"fromAccountIndex"`
		FromEthAddr      apitypes.HezEthAddr     `json:"fromHezEthereumAddress"`
		FromBJJ          apitypes.HezBJJ         `json:"fromBJJ"`
		TxID             hezCommon.TxID          `json:"id"`
		BatchNum         hezCommon.BatchNum      `json:"batchNum"`
		L1orL2           string                  `json:"L1orL2"`
		L1Info           interface{}             `json:"L1Info"`
		L2Info           interface{}             `json:"L2Info"`
		Nonce            hezCommon.Nonce         `json:"nonce"`
		RequestAmount    apitypes.BigIntStr      `json:"requestAmount"`
		RequestFee       hezCommon.FeeSelector   `json:"requestFee"`
		RequestFromIdx   StrHezIdx               `json:"requestFromAccountIndex"`
		RequestNonce     hezCommon.Nonce         `json:"requestNonce"`
		RequestToIdx     hezCommon.Idx           `json:"requestToAccountIndex"`
		RequestToBJJ     babyjub.PublicKeyComp   `json:"requestToBJJ"`
		RequestToEthAddr ethCommon.Address       `json:"requestToHezEthereumAddress"`
		RequestTokenID   hezCommon.TokenID       `json:"requestTokenId"`
		Signature        string                  `json:"signature"`
		State            hezCommon.PoolL2TxState `json:"state"`
		Timestamp        time.Time               `json:"timestamp"`
		ToIdx            StrHezIdx               `json:"toAccountIndex"`
		ToEthAddr        apitypes.HezEthAddr     `json:"toHezEthereumAddress"`
		ToBJJ            apitypes.HezBJJ         `json:"toBjj"`
		Token            hezCommon.Token         `json:"token"`
		Type             hezCommon.TxType        `json:"type"`
	}

	// TxAPI is a representation of a tx history API response.
	TxAPI struct {
		Txs          []TxHistory `json:"transactions"`
		PendingItems uint64      `json:"pendingItems"`
	}

	// BatchAPI is a representation of a batches API response.
	BatchAPI struct {
		Batches      []Batch `json:"batches"`
		PendingItems uint64  `json:"pendingItems"`
	}

	// TokenAPI is a representation of a tokens API response.
	TokenAPI struct {
		Tokens       Tokens `json:"tokens"`
		PendingItems uint64 `json:"pendingItems"`
	}

	// Tokens is a representation of a list of tokens.
	Tokens []hezCommon.Token

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

	// AccountAuth is a representation of a account authentication API request.
	AccountAuth struct {
		EthAddr   string `json:"hezEthereumAddress"`
		Bjj       string `json:"bjj"`
		Signature string `json:"signature"`
	}

	// CreateAccountAuthAPI is a representation of a account authentication API response.
	CreateAccountAuthAPI struct {
		Message string `json:"Message"`
	}

	// AccountAuthAPI is a representation of a account authentication API response.
	AccountAuthAPI struct {
		EthAddr   string                `json:"hezEthereumAddress"`
		Bjj       string                `json:"bjj"`
		Signature apitypes.EthSignature `json:"signature"`
		Timestamp time.Time             `json:"timestamp"`
		Message   string                `json:"Message"`
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

// GetFirstAccount get the first account by token ID
func (acs *Accounts) GetFirstAccount(tokenID hezCommon.TokenID) (Account, error) {
	for _, ac := range *acs {
		if ac.Token.TokenID == tokenID {
			return ac, nil
		}
	}
	return Account{}, errors.E("account not found",
		errors.Params{"token_id": tokenID})
}

// GetToken get a token by symbol
func (ts *Tokens) GetToken(symbol string) (hezCommon.Token, error) {
	for _, t := range *ts {
		if t.Symbol == symbol {
			return t, nil
		}
	}
	return hezCommon.Token{}, errors.E("token not supported",
		errors.Params{"symbol": symbol})
}
