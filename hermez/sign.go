package hermez

import (
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/Pantani/errors"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// CreateTx create a Hermez transaction
func CreateTx(chainID uint16, to string, amount *big.Int, pk babyjub.PrivateKey,
	fromIdx hezCommon.Idx, tokenID hezCommon.TokenID, nonce hezCommon.Nonce,
	fee hezCommon.FeeSelector) (*hezCommon.PoolL2Tx, error) {

	toBjj, err := HezStrToBJJ(to)
	if err != nil {
		return nil, err
	}

	// Generate tx
	tx := &hezCommon.PoolL2Tx{
		FromIdx:   fromIdx,
		ToBJJ:     toBjj,
		ToEthAddr: hezCommon.FFAddr,
		Amount:    amount,
		Fee:       fee,
		TokenID:   tokenID,
		Nonce:     nonce,
		Type:      hezCommon.TxTypeTransferToBJJ,
	}

	// Set tx type and id
	tx, err = hezCommon.NewPoolL2Tx(tx)
	if err != nil {
		return nil, err
	}

	// Sign tx
	toSign, err := tx.HashToSign(chainID)
	if err != nil {
		return nil, err
	}
	sig := pk.SignPoseidon(toSign)
	tx.Signature = sig.Compress()
	return tx, nil
}

// HezStrToBJJ convert bjj public key to babyjub.PublicKeyComp
func HezStrToBJJ(s string) (babyjub.PublicKeyComp, error) {
	const decodedLen = 33
	const encodedLen = 44
	formatErr := errors.E("invalid BJJ format. Must follow this regex: ^hez:[A-Za-z0-9_-]{44}$", errors.Params{"bjj": s})
	encoded := strings.TrimPrefix(s, "hez:")
	if len(encoded) != encodedLen {
		return hezCommon.EmptyBJJComp, formatErr
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return hezCommon.EmptyBJJComp, formatErr
	}
	if len(decoded) != decodedLen {
		return hezCommon.EmptyBJJComp, formatErr
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return hezCommon.EmptyBJJComp, errors.E("checksum verification failed")
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	return bjjComp, nil
}
