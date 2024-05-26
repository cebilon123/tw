package ethereum

import (
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const checkBlockNumberIntervalSeconds = 5

var (
	getCurrentBlockFunc         = getCurrentBlock
	getTransactionsForBlockFunc = getTransactionsForBlock
)

type JSONRpcBasedObserver struct {
	httpClient *http.Client
	logger     *log.Logger

	closeChan chan struct{}
}

var _ Observer = (*JSONRpcBasedObserver)(nil)
var _ io.Closer = (*JSONRpcBasedObserver)(nil)

func NewJSONRpcBasedObserver(httpClient *http.Client, logger *log.Logger) *JSONRpcBasedObserver {
	return &JSONRpcBasedObserver{
		httpClient: httpClient,
		logger:     logger,
		closeChan:  make(chan struct{}, 1),
	}
}

// ObserveAddress observes blockchain for changes to a given address transactions, if any found
// it returns that transaction on the channel.
func (j *JSONRpcBasedObserver) ObserveAddress(address string) (<-chan Transaction, error) {
	// we are going to check if new block appeared
	blockNumberChan := make(chan string)
	transactionsChan := make(chan Transaction)

	var mu sync.Mutex
	closed := false

	// if observer is closed we are closing the transaction chan
	go func() {
		<-j.closeChan

		mu.Lock()

		defer mu.Unlock()

		closed = true

		close(transactionsChan)
		close(blockNumberChan)
	}()

	// we are checking if new blocks appeared
	go func() {
		for {
			num, err := getCurrentBlockFunc(j.httpClient)
			if err != nil {
				j.logger.Printf("get current block error: %s", err.Error())
			}

			blockNumberChan <- num

			time.Sleep(time.Second * checkBlockNumberIntervalSeconds)
		}
	}()

	go func() {
		var lastBlockNum int64
		for num := range blockNumberChan {
			n := new(big.Int)
			// passing 0, it will pick base based on the string
			n.SetString(num, 0)
			currentBlockNum := n.Int64()

			if lastBlockNum == 0 {
				lastBlockNum = currentBlockNum
				continue
			}

			// if there is no dif in block num it means there are no new transactions
			dif := currentBlockNum - lastBlockNum
			if dif == 0 {
				continue
			}

			// for each new block after the last block we are fetching the transactions
			// and then we are checking if there are any for given address
			for i := range dif {
				blockNum := lastBlockNum + i
				transactions, err := getTransactionsForBlockFunc(j.httpClient, fmt.Sprintf("%x", blockNum))
				if err != nil {
					j.logger.Printf("get transactions for block error: %s", err.Error())
					continue
				}

				for _, transaction := range transactions {

					// there is an transaction for a given address, we are sending it to chan
					if transaction.From == address || transaction.To == address {
						// we want to be sure that we are not going to send anything more on the closed channel
						mu.Lock()
						if closed {
							mu.Unlock()
							return
						}

						transactionsChan <- transaction

						mu.Unlock()
					}
				}
			}

			lastBlockNum = currentBlockNum
		}
	}()

	return transactionsChan, nil
}

func (j *JSONRpcBasedObserver) Close() error {
	j.closeChan <- struct{}{}

	return nil
}
