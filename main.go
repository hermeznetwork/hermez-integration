package main

import (
	"context"
	"encoding/hex"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
	"github.com/hermeznetwork/hermez-integration/hermez"
	"github.com/hermeznetwork/hermez-integration/track"
	"github.com/hermeznetwork/hermez-integration/transaction"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"golang.org/x/sync/errgroup"
)

const (
	// Rinkeby chain id
	chainID         = uint16(4)
	nodeURL         = "https://api.testnet.hermez.io"
	poolingInterval = 10 * time.Second
)

func main() {
	err := run(nodeURL, chainID, poolingInterval)
	if err != nil {
		logger.Fatal(err)
	}
}

func run(nodeURL string, chainID uint16, poolingInterval time.Duration) error {
	// init context
	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx := errgroup.WithContext(ctx)
	defer cancel()

	// create a new Hermez node client
	c := client.New(nodeURL)

	// get supported tokens
	tokens, err := c.GetTokens()
	if err != nil {
		return err
	}
	for _, t := range tokens.Tokens {
		logger.Info("Token "+t.Name, logger.Params{
			"TokenID":     t.TokenID,
			"Decimals":    t.Decimals,
			"EthAddr":     t.EthAddr,
			"EthBlockNum": t.EthBlockNum,
			"Symbol":      t.Symbol,
		})
	}
	ethToken, err := tokens.Tokens.GetToken("ETH")
	if err != nil {
		return err
	}

	// represents the mnemonic for the exchange user wallets
	exchangeMnemonic := "lava dinosaur defy stone aim faint suspect harsh ranch sorry network wrestle"
	numberOfUsers := 10
	ethUserWallets := make([]string, 0)
	bjjUserWallets := make([]string, 0)

	// Increase the wallet index to generate a new wallet based
	// in the bip39, starting from zero
	for walletIndex := 0; walletIndex < numberOfUsers; walletIndex++ {
		bjj, err := hermez.NewBJJ(exchangeMnemonic, walletIndex)
		if err != nil {
			return err
		}
		pkBuf := [hermez.PkLength]byte(bjj.PrivateKey)
		logger.Info("User Wallet Create", logger.Params{
			"index":           walletIndex,
			"hez_eth_address": bjj.HezEthAddress,
			"hez_bjj_address": bjj.HezBjjAddress,
			"private_key":     "0x" + hex.EncodeToString(pkBuf[:]),
		})
		ethUserWallets = append(ethUserWallets, bjj.HezEthAddress)
		bjjUserWallets = append(bjjUserWallets, bjj.HezBjjAddress)
	}

	// track incoming track
	grp.Go(track.Deposits(c, ethUserWallets, bjjUserWallets, poolingInterval))

	// represents the mnemonic for outside wallet, it is assumed that
	// the user has already Ether in Hermez Network
	outWalletMnemonic := "butter embrace sunny tilt soap where soul finish shop west rough flock"
	// get the first index wallet from the mnemonic (https://iancoleman.io/bip39)
	walletIndex := 0

	// Create a baby jubjub wallet based in the mnemonic and index
	bjj, err := hermez.NewBJJ(outWalletMnemonic, walletIndex)
	if err != nil {
		return err
	}
	pkBuf := [hermez.PkLength]byte(bjj.PrivateKey)
	logger.Info("Out Wallet Create", logger.Params{
		"hez_eth_address": bjj.HezEthAddress,
		"hez_bjj_address": bjj.HezBjjAddress,
		"private_key":     "0x" + hex.EncodeToString(pkBuf[:]),
	})

	// A fee is a percentage value from the token amount, and the fee amount in USD must
	// be greater than the minimum fee value the coordinator accepts. The fee value in the
	// L2 transaction apply a factor encoded by an index from the transaction fee table:
	// https://docs.hermez.io/#/developers/protocol/hermez-protocol/fee-table?id=transaction-fee-table
	amount := big.NewInt(6000000000000000)
	fee := hezCommon.FeeSelector(126) // 10.2%
	feeAmount, err := hezCommon.CalcFeeAmount(amount, fee)
	if err != nil {
		return err
	}
	hashes := make([]string, 0)

	logger.Info("Fee", logger.Params{
		"amount_wei":     amount.String(),
		"amount_eth":     hermez.WeiToEther(amount).String(),
		"fee_selector":   fee,
		"fee_percentage": fee.Percentage(),
		"fee_amount_wei": feeAmount.String(),
		"fee_amount_eth": hermez.WeiToEther(feeAmount).String(),
	})

	// Create a transfer to the first baby jubjub user address
	time.Sleep(3 * time.Second)
	rand.Seed(time.Now().Unix())
	bjjIndex := rand.Intn(len(bjjUserWallets))
	toHezBjjAddr := bjjUserWallets[bjjIndex]
	txID, err := transaction.TransferToBjj(bjj, c, chainID, toHezBjjAddr, amount, fee, ethToken)
	if err != nil {
		return err
	}
	hashes = append(hashes, txID)
	logger.Info("transferToBjj", logger.Params{"tx_id": txID})

	// Create a transfer to the second user ethereum address
	time.Sleep(3 * time.Second)
	rand.Seed(time.Now().Unix())
	ethIndex := rand.Intn(len(ethUserWallets))
	toHezEthAddr := ethUserWallets[ethIndex]
	toHezEthAddr = strings.Replace(toHezEthAddr, "hez:", "", -1)
	txID, err = transaction.TransferToEthAddress(bjj, c, chainID, toHezEthAddr, amount, fee, ethToken)
	if err != nil {
		return err
	}
	hashes = append(hashes, txID)
	logger.Info("transferToEthAddress", logger.Params{"tx_id": txID})

	// Create a transfer to an already existing account into the
	// network using the idx (Merkle tree index).
	toHezAddr := "hez:0xd9391B20559777E1b94954Ed84c28541E35bFEb8"
	toIdx, _, err := transaction.GetAccountInfo(c, nil, &toHezAddr, amount, ethToken.TokenID)
	if err != nil {
		return err
	}
	txID, err = transaction.Transfer(bjj, c, chainID, toIdx, amount, fee, ethToken)
	if err != nil {
		return err
	}
	hashes = append(hashes, txID)
	logger.Info("transfer to idx", logger.Params{"tx_id": txID})

	// Transfer tokens from an account to the exit tree, L2 --> L1
	time.Sleep(3 * time.Second)
	txID, err = transaction.Exit(bjj, c, chainID, amount, fee, ethToken)
	if err != nil {
		return err
	}
	logger.Info("exit", logger.Params{"tx_id": txID})

	// track transactions
	grp.Go(track.Txs(c, hashes, poolingInterval))

	// wait for SIGINT/SIGTERM.
	waiter := make(chan os.Signal, 1)
	signal.Notify(waiter, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-waiter:
		logger.Info("waiter signal", logger.Params{"signal": sig})
		cancel()
	case <-ctx.Done():
		logger.Info("context done")
		if err := ctx.Err(); err != nil {
			return errors.E("context error", err)
		}
	}
	if err := grp.Wait(); err != nil {
		return errors.E("error group failure", err)
	}
	logger.Info("Exiting gracefully")
	return nil
}
