package client

import (
	"encoding/base64"
	"strconv"

	ethCommon "github.com/ethereum/go-ethereum/common"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// NewTxRequest convert L2 tx to API request model
func NewTxRequest(poolTx hezCommon.PoolL2Tx, token hezCommon.Token) *Tx {
	toIdx := ""
	if poolTx.ToIdx != 0 {
		toIdx = idxToHez(poolTx.ToIdx, token.Symbol)
	}
	toEth := ""
	if poolTx.ToEthAddr != hezCommon.EmptyAddr {
		toEth = ethAddrToHez(poolTx.ToEthAddr)
	}
	toBJJ := bjjToString(poolTx.ToBJJ)
	if poolTx.ToBJJ != hezCommon.EmptyBJJComp {
		toBJJ = bjjToString(poolTx.ToBJJ)
	}
	return &Tx{
		TxID:      poolTx.TxID,
		ToIdx:     toIdx,
		ToEthAddr: toEth,
		ToBJJ:     toBJJ,
		Type:      poolTx.Type,
		TokenID:   poolTx.TokenID,
		FromIdx:   idxToHez(poolTx.FromIdx, token.Symbol),
		Amount:    poolTx.Amount.String(),
		Fee:       poolTx.Fee,
		Nonce:     poolTx.Nonce,
		Signature: poolTx.Signature,
	}
}

// idxToHez convert idx to hez idx
func idxToHez(idx hezCommon.Idx, tokenSymbol string) string {
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}

// idxToHez convert eth address to hez address
func ethAddrToHez(addr ethCommon.Address) string {
	return "hez:" + addr.String()
}

// bjjToString convert the BJJ public key to string
func bjjToString(bjj babyjub.PublicKeyComp) string {
	pkComp := [32]byte(bjj)
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}
