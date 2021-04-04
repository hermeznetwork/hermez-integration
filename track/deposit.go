package track

import (
	"time"

	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
)

// Deposits get last batch number and get the transactions
func Deposits(c *client.Client, ethAddr, bjjAddr []string, interval time.Duration) func() error {
	return func() error {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				// Fetch the last batch for we can pulling the transactions
				lastBatch, err := c.GetLastBatch()
				if err != nil {
					return err
				}
				logger.Info("Last Batch", logger.Params{"last_batch": lastBatch.BatchNum})

				// Get all transactions for a batch for tracking the track
				batch, err := c.GetBatchTxs(lastBatch.BatchNum)
				if err != nil {
					return err
				}
				logger.Info("Batch", logger.Params{"batch": lastBatch.BatchNum, "txs": len(batch.Txs)})

				for _, tx := range batch.Txs {
					toEthAddr, err := tx.ToEthAddr.ToEthAddr()
					if err != nil {
						return err
					}
					if contains(ethAddr, toEthAddr.String()) {
						logger.Info("New tx found",
							logger.Params{"batch": lastBatch.BatchNum,
								"tx": tx.TxID, "eth_addr": toEthAddr.String()})
						continue
					}
					toBjjAddr, err := tx.ToBJJ.ToBJJ()
					if err != nil {
						return err
					}
					if contains(bjjAddr, toBjjAddr.String()) {
						logger.Info("New tx found",
							logger.Params{"batch": lastBatch.BatchNum,
								"tx": tx.TxID, "bjjAddr": toBjjAddr.String()})
						continue
					}
				}
			}
		}
	}
}

// contains check if a string contains into a slice
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
