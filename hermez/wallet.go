package hermez

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Pantani/errors"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	// PkLength represents the size of the baby jubjub private key
	PkLength uint16 = 32

	// wallet configs
	msg               = "Hermez Network account access.\n\nSign this message if you are in a trusted application only."
	ethDerivationPath = "m/44'/60'/0'/0/%d"
)

// Wallet represents a wallet object with a private key,
// public key and a baby jubjub hez address
type Wallet struct {
	PrivateKey babyjub.PrivateKey
	PublicKey  babyjub.PublicKeyComp
	HezAddress string
}

// NewBJJ create a baby jubjub address from the mnemonic
// and the derivation path. It returns a wallet object
// with a private key, public key and a baby jubjub hez
// address and a error if occurs.
func NewBJJ(mnemonic string, index int) (*Wallet, error) {
	var sk babyjub.PrivateKey
	var pk babyjub.PublicKeyComp
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

	// Sign message
	signature, err := w.SignText(ethAccount, []byte(msg))
	if err != nil {
		return nil, errors.E("Signing message error", err)
	}

	signature[len(signature)-1] += 27

	// Hash signature
	var sb strings.Builder
	sb.WriteString("0x")
	sb.WriteString(hex.EncodeToString(signature))
	hash := ethCrypto.Keccak256([]byte(sb.String()))

	copy(sk[:], hash[:])

	// Create the Baby Jubjub hez address
	compressPk := sk.Public().Compress().String()
	b, err := hex.DecodeString(compressPk)
	if err != nil {
		return nil, errors.E("Decode pk error", err)
	}
	hezAddress := NewHezBJJ(sk.Public().Compress())

	// Create the Hermez hez address
	pkBytes := hezCommon.SwapEndianness(b)
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkBytes[:])

	return &Wallet{
		PrivateKey: sk,
		PublicKey:  pk,
		HezAddress: hezAddress,
	}, nil
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
