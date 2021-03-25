package hermez

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
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

// CreateBJJFromEthAddr create a baby jubjub address from the mnemonic and the derivation path
func CreateBJJFromEthAddr(mnemonic string, derivationPath int) (babyjub.PrivateKey, error) {
	var sk babyjub.PrivateKey
	w, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return sk, err
	}

	// Generate ETH account
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf(ethDerivationPath, derivationPath))
	ethAccount, err := w.Derive(path, true)
	if err != nil {
		return sk, err
	}

	// Sign message
	signature, err := w.SignText(ethAccount, []byte(msg))
	if err != nil {
		return sk, errors.New("Error signing: " + err.Error())
	}

	signature[len(signature)-1] += 27
	// Hash signature
	var sb strings.Builder
	sb.WriteString("0x")
	sb.WriteString(hex.EncodeToString(signature))
	hash := ethCrypto.Keccak256([]byte(sb.String()))

	copy(sk[:], hash[:])
	return sk, nil
}
