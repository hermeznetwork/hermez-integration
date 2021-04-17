package main

import (
	"context"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hermeznetwork/hermez-integration/client"
	"github.com/hermeznetwork/hermez-integration/hermez"
	"github.com/hermeznetwork/hermez-integration/track"
	"github.com/hermeznetwork/hermez-integration/transaction"
	hezCommon "github.com/hermeznetwork/hermez-node/common"
	"golang.org/x/sync/errgroup"
)

const (
	// chainID represents the Rinkeby chain id
	chainID = uint16(4)
	// nodeURL represents the Rinkeby testnet node URL
	nodeURL = "https://api.testnet.hermez.io"
	// rollupContract represents the Rinkeby rollup contract address
	rollupContract = "0x679b11E0229959C1D3D27C9d20529E4C5DF7997c"
	// poolingInterval pooling interval to check the transactions state
	poolingInterval = 10 * time.Second
)

func main() {
	err := run(nodeURL, rollupContract, chainID, poolingInterval)
	if err != nil {
		logger.Fatal(err)
	}
}

func run(nodeURL, rollupContract string, chainID uint16, poolingInterval time.Duration) error {
	// init context
	ctx, cancel := context.WithCancel(context.Background())
	grp, gctx := errgroup.WithContext(ctx)
	defer cancel()

	contract := ethCommon.HexToAddress(rollupContract)

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
	startUserIndex := 30
	numberOfUsers := 13
	ethUserWallets := make([]string, 0)
	bjjUserWallets := make([]string, 0)

	// Increase the wallet index to generate a new wallet based
	// in the bip39, starting from zero
	for walletIndex := startUserIndex; walletIndex < startUserIndex+numberOfUsers; walletIndex++ {
		bjj, err := hermez.NewBJJ(exchangeMnemonic, walletIndex, chainID, contract)
		if err != nil {
			return err
		}
		pkBuf := [hermez.PkLength]byte(bjj.PrivateKey)
		logger.Info("User Wallet", logger.Params{
			"index":           walletIndex,
			"hez_eth_address": bjj.HezEthAddress,
			"hez_bjj_address": bjj.HezBjjAddress,
			"private_key":     hexutil.Encode(pkBuf[:]),
		})
		ethUserWallets = append(ethUserWallets, bjj.HezEthAddress)
		bjjUserWallets = append(bjjUserWallets, bjj.HezBjjAddress)

		// Get the signature from the hez eth address
		_, err = c.AccountAuth(bjj.HezEthAddress)
		if err != nil {
			// If the signature not exist, create a new one
			err = c.AccountCreationAuth(bjj.HezEthAddress, bjj.HezBjjAddress, bjj.Signature)
			if err != nil {
				return err
			}
		}
		logger.Info("User account authentication created", logger.Params{
			"hez_eth_address": bjj.HezEthAddress,
			"signature":       bjj.Signature,
		})
	}

	// track incoming track
	grp.Go(track.Deposits(c, ethUserWallets, bjjUserWallets, poolingInterval))

	// represents the mnemonic for outside wallet, it is assumed that
	// the user has already Ether in Hermez Network
	outWalletMnemonic := "butter embrace sunny tilt soap where soul finish shop west rough flock"
	// get the first index wallet from the mnemonic (https://iancoleman.io/bip39)
	outWalletIndex := 0

	// Create a baby jubjub wallet based in the mnemonic and index
	bjj, err := hermez.NewBJJ(outWalletMnemonic, outWalletIndex, chainID, contract)
	if err != nil {
		return err
	}
	pkBuf := [hermez.PkLength]byte(bjj.PrivateKey)
	logger.Info("Out Wallet", logger.Params{
		"hez_eth_address": bjj.HezEthAddress,
		"hez_bjj_address": bjj.HezBjjAddress,
		"private_key":     hexutil.Encode(pkBuf[:]),
	})

	// Get the signature from the hez eth address
	_, err = c.AccountAuth(bjj.HezEthAddress)
	if err != nil {
		// If the signature not exist, create a new one
		err = c.AccountCreationAuth(bjj.HezEthAddress, bjj.HezBjjAddress, bjj.Signature)
		if err != nil {
			return err
		}
	}
	logger.Info("User account authentication created", logger.Params{
		"hez_eth_address": bjj.HezEthAddress,
		"signature":       bjj.Signature,
	})

	// A fee is a percentage value from the token amount, and the fee amount in USD must
	// be greater than the minimum fee value the coordinator accepts. The fee value in the
	// L2 transaction apply a factor encoded by an index from the transaction fee table:
	// https://docs.hermez.io/#/developers/protocol/hermez-protocol/fee-table?id=transaction-fee-table
	amount := big.NewInt(5920000000000000)
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
	// Get account idx, nonce and check the balance
	fromIdx, nonce, err := transaction.GetAccountInfo(c, &bjj.HezBjjAddress, nil, ethToken.TokenID)
	if err != nil {
		return err
	}

	// Create a transfer to the first baby jubjub user address
	time.Sleep(3 * time.Second)
	rand.Seed(time.Now().Unix())
	bjjIndex := rand.Intn(len(bjjUserWallets))
	toHezBjjAddr := bjjUserWallets[bjjIndex]
	txID, err := transaction.TransferToBjj(bjj, c, chainID, fromIdx, toHezBjjAddr, amount, fee, ethToken, nonce)
	if err != nil {
		return err
	}
	hashes = append(hashes, txID)
	logger.Info("transferToBjj", logger.Params{"tx_id": txID})

	// Create a transfer to the second user ethereum address
	nonce++
	time.Sleep(3 * time.Second)
	rand.Seed(time.Now().Unix())
	ethIndex := rand.Intn(len(ethUserWallets))
	toHezAddr := ethUserWallets[ethIndex]
	txID, err = transaction.TransferToEthAddress(bjj, c, chainID, fromIdx, toHezAddr, amount, fee, ethToken, nonce)
	if err != nil {
		return err
	}
	hashes = append(hashes, txID)
	logger.Info("transferToEthAddress", logger.Params{"tx_id": txID})

	// Create a transfer to an already existing account into the
	// network using the idx (Merkle tree index).
	nonce++
	toHezAddr = "hez:0xbA00D84Ddbc8cAe67C5800a52496E47A8CaFcd27"
	toIdx, _, err := transaction.GetAccountInfo(c, nil, &toHezAddr, ethToken.TokenID)
	if err != nil {
		return err
	}
	txID, err = transaction.Transfer(bjj, c, chainID, fromIdx, toIdx, amount, fee, ethToken, nonce)
	if err != nil {
		return err
	}
	hashes = append(hashes, txID)
	logger.Info("transfer to idx", logger.Params{"tx_id": txID})

	//Transfer tokens from an account to the exit tree, L2 --> L1
	nonce++
	time.Sleep(3 * time.Second)
	txID, err = transaction.Exit(bjj, c, chainID, fromIdx, amount, fee, ethToken, nonce)
	if err != nil {
		return err
	}
	logger.Info("exit", logger.Params{"tx_id": txID})

	// track transactions
	grp.Go(track.Txs(c, hashes, poolingInterval))

	// wait for SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)
	select {
	case <-quit:
		cancel()
		logger.Info("Exiting gracefully")
		if err := grp.Wait(); err != nil {
			return errors.E("error group failure", err)
		}
	case <-gctx.Done():
		logger.Info("context done")
		return gctx.Err()
	}
	return nil
}
