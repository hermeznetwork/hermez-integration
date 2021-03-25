package hermez

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
)

var (
	// EthToken is a representation of a ETH Token on Hermez Network
	EthToken = hezCommon.Token{
		TokenID:     0,
		Name:        "Ether",
		Symbol:      "ETH",
		Decimals:    18,
		EthBlockNum: 0,
		EthAddr:     ethCommon.BigToAddress(big.NewInt(0)),
	}
)
