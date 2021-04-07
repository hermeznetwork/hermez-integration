package track

import (
	"time"

	"github.com/Pantani/logger"
	"github.com/hermeznetwork/hermez-integration/client"
)

// Txs track if transaction was forged
func Txs(c *client.Client, hashes []string, interval time.Duration) func() error {
	return func() error {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				for i, hash := range hashes {
					poolTx, err := c.GetPoolTx(hash)
					if err == nil && poolTx != nil && poolTx.TxID.String() == hash {
						logger.Info("Tx stills on pool", logger.Params{"tx_id": poolTx.TxID})
						continue
					}

					tx, err := c.GetTx(hash)
					if err == nil && tx != nil && tx.TxID.String() == hash {
						logger.Info("Tx was forged", logger.Params{"tx_id": tx.TxID.String()})
						hashes = append(hashes[:i], hashes[i+1:]...)
						continue
					}
				}
				if len(hashes) == 0 {
					return nil
				}
			}
		}
	}
}
