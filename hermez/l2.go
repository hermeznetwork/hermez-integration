package hermez

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// CreateTransfer create a L2 transfer to baby jubjub transaction
func CreateTransfer(chainID uint16, toIdx hezCommon.Idx, amount *big.Int, privateKey babyjub.PrivateKey,
	fromIdx hezCommon.Idx, tokenID hezCommon.TokenID, nonce hezCommon.Nonce,
	fee hezCommon.FeeSelector) (*hezCommon.PoolL2Tx, error) {

	return createTxObject(chainID, hezCommon.EmptyBJJComp, hezCommon.FFAddr,
		amount, privateKey, fromIdx, toIdx,
		tokenID, nonce, fee, hezCommon.TxTypeTransfer)
}

// CreateTransferToBjj create a L2 transfer to baby jubjub transaction
func CreateTransferToBjj(chainID uint16, to string, amount *big.Int, privateKey babyjub.PrivateKey,
	fromIdx hezCommon.Idx, tokenID hezCommon.TokenID, nonce hezCommon.Nonce,
	fee hezCommon.FeeSelector) (*hezCommon.PoolL2Tx, error) {

	toBjj, err := HezStrToBJJ(to)
	if err != nil {
		return nil, err
	}

	return createTxObject(chainID, toBjj, hezCommon.FFAddr,
		amount, privateKey, fromIdx, hezCommon.Idx(0),
		tokenID, nonce, fee, hezCommon.TxTypeTransferToBJJ)
}

// CreateTransferToEthAddress create a L2 transfer to eth address transaction
func CreateTransferToEthAddress(chainID uint16, to string, amount *big.Int, privateKey babyjub.PrivateKey,
	fromIdx hezCommon.Idx, tokenID hezCommon.TokenID, nonce hezCommon.Nonce,
	fee hezCommon.FeeSelector) (*hezCommon.PoolL2Tx, error) {

	toEthAddr := ethCommon.HexToAddress(to)
	return createTxObject(chainID, hezCommon.EmptyBJJComp, toEthAddr,
		amount, privateKey, fromIdx, hezCommon.Idx(0),
		tokenID, nonce, fee, hezCommon.TxTypeTransferToEthAddr)
}

// CreateExit create a L2 exit transaction
func CreateExit(chainID uint16, amount *big.Int, privateKey babyjub.PrivateKey,
	fromIdx hezCommon.Idx, tokenID hezCommon.TokenID, nonce hezCommon.Nonce,
	fee hezCommon.FeeSelector) (*hezCommon.PoolL2Tx, error) {

	return createTxObject(chainID, hezCommon.EmptyBJJComp, hezCommon.FFAddr,
		amount, privateKey, fromIdx, hezCommon.Idx(1),
		tokenID, nonce, fee, hezCommon.TxTypeExit)
}

// createTxObject create and validate the transaction object
func createTxObject(chainID uint16, toBjj babyjub.PublicKeyComp, toEthAddr ethCommon.Address,
	amount *big.Int, privateKey babyjub.PrivateKey, fromIdx, toIdx hezCommon.Idx, tokenID hezCommon.TokenID,
	nonce hezCommon.Nonce, fee hezCommon.FeeSelector, txType hezCommon.TxType) (*hezCommon.PoolL2Tx, error) {

	// Create the l2 tx object
	tx := &hezCommon.PoolL2Tx{
		FromIdx:   fromIdx,
		ToBJJ:     toBjj,
		ToEthAddr: toEthAddr,
		ToIdx:     toIdx,
		Amount:    amount,
		Fee:       fee,
		TokenID:   tokenID,
		Nonce:     nonce,
		Type:      txType,
	}

	// Set tx type and id
	tx, err := hezCommon.NewPoolL2Tx(tx)
	if err != nil {
		return nil, err
	}

	// Sign tx
	toSign, err := tx.HashToSign(chainID)
	if err != nil {
		return nil, err
	}
	sig := privateKey.SignPoseidon(toSign)
	tx.Signature = sig.Compress()
	return tx, nil
}
