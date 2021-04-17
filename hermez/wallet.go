package hermez

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/Pantani/errors"
	"github.com/ethereum/go-ethereum/accounts"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	// PkLength represents the size of the baby jubjub private key
	PkLength uint16 = 32
	// msg message to be signed and generate the BJJ
	msg = "Hermez Network account access.\n\nSign this message if you are in a trusted application only."
	// ethDerivationPath represents the ethereum bip-44 derivation path
	ethDerivationPath = "m/44'/60'/0'/0/%d"
)

type (
	// Wallet represents a wallet object with a private key,
	// public key and a baby jubjub hez address
	Wallet struct {
		PrivateKey    babyjub.PrivateKey
		PublicKey     babyjub.PublicKeyComp
		HezBjjAddress string
		HezEthAddress string
		Signature     string
	}
)

// NewBJJ create a baby jubjub address from the mnemonic
// and the derivation path. It returns a wallet object
// with a private key, public key and a baby jubjub hez
// address and a error if occurs.
func NewBJJ(mnemonic string, index int, chainID uint16,
	rollupContract ethCommon.Address) (*Wallet, error) {
	w, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, errors.E("New wallet error", err)
	}

	// Generate ETH account
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf(ethDerivationPath, index))
	ethAccount, err := w.Derive(path, true)
	if err != nil {
		return nil, errors.E("Path derivation error", err)
	}
	hezEthAddress := "hez:" + ethAccount.Address.String()

	// Sign message
	signature, err := w.SignText(ethAccount, []byte(msg))
	if err != nil {
		return nil, errors.E("Signing message error", err)
	}

	signature[len(signature)-1] += 27
	sigEncoded := hexutil.Encode(signature)
	hash := ethCrypto.Keccak256([]byte(sigEncoded))

	var sk babyjub.PrivateKey
	copy(sk[:], hash[:])

	// Create the Baby Jubjub hez address
	compressPk := sk.Public().Compress().String()
	b, err := hex.DecodeString(compressPk)
	if err != nil {
		return nil, errors.E("Decode compress pk error", err)
	}
	hezBjjAddress := NewHezBJJ(sk.Public().Compress())

	// Create the Hermez hez address
	pkBytes := hezCommon.SwapEndianness(b)
	var pk babyjub.PublicKeyComp
	copy(pk[:], pkBytes[:])

	// Create the wallet authentication signature
	// https://docs.hermez.io/#/developers/protocol/hermez-protocol/protocol?id=regular-rollup-account
	// Use the endianness not swapped because the method will swap it again
	var sigPk babyjub.PublicKeyComp
	copy(sigPk[:], b[:])
	authSign, err := createSignature(w, ethAccount, sigPk, chainID, rollupContract)
	if err != nil {
		return nil, errors.E("fail to generate the account authentication", err)
	}

	return &Wallet{
		PrivateKey:    sk,
		PublicKey:     pk,
		HezBjjAddress: hezBjjAddress,
		HezEthAddress: hezEthAddress,
		Signature:     authSign,
	}, nil
}

// createSignature creates the wallet authentication signature
func createSignature(wallet *hdwallet.Wallet, ethAccount accounts.Account,
	pk babyjub.PublicKeyComp, chainID uint16, rollupContract ethCommon.Address) (string, error) {
	ethSk, err := wallet.PrivateKey(ethAccount)
	if err != nil {
		return "", err
	}

	auth := &hezCommon.AccountCreationAuth{
		EthAddr: ethAccount.Address,
		BJJ:     pk,
	}
	err = auth.Sign(func(hash []byte) ([]byte, error) {
		return ethCrypto.Sign(hash, ethSk)
	}, chainID, rollupContract)

	hash, err := auth.HashToSign(chainID, rollupContract)
	if err != nil {
		return "", err
	}

	signature, err := ethCrypto.Sign(hash, ethSk)
	if err != nil {
		return "", err
	}
	signature[64] += 27

	if !auth.VerifySignature(chainID, rollupContract) {
		return "", errors.E("invalid signature")
	}
	return hexutil.Encode(signature), nil
}

// NewHezBJJ creates a HezBJJ from a *babyjub.PublicKeyComp.
// Calling this method with a nil bjj causes panic
func NewHezBJJ(pkComp babyjub.PublicKeyComp) string {
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

// HezStrToBJJ convert bjj public key to babyjub.PublicKeyComp
func HezStrToBJJ(s string) (babyjub.PublicKeyComp, error) {
	const (
		decodedLen = 33
		encodedLen = 44
	)
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

// WeiToEther converts a wei value (*big.Int) to a ether value (*big.Float)
func WeiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236)
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236)
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.Ether))
}
