package ethereum

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"tw/internal/clogger"
)

func TestJSONRpcBasedObserver_ObserveAddress_Then_GetTransactions(t *testing.T) {
	currBlockNum := 0

	var expectedTransactions []Transaction

	transactionNum := 0

	apiWrapper := &mockApiWrapper{
		getCurrentBlockFunc: func(httpClient *http.Client) (string, error) {
			currBlockNum++

			return fmt.Sprintf("%x", currBlockNum), nil
		},
		getTransactionsForBlockFunc: func(httpClient *http.Client, blockNum string) ([]Transaction, error) {
			transactions := []Transaction{
				{
					From: "test",
					To:   fmt.Sprintf("to%d", transactionNum+1),
				},
			}

			expectedTransactions = append(expectedTransactions, transactions...)

			transactionNum++

			return transactions, nil
		},
	}

	observer := JSONRpcBasedObserver{
		httpClient: http.DefaultClient,
		closeChan:  make(chan struct{}, 1),
		logger:     clogger.ConsoleLogger,
		apiWrapper: apiWrapper,
	}

	go func() {
		for {
			// let some transactions go through
			if transactionNum >= 2 {
				_ = observer.Close()
			}
		}
	}()

	transactionsChan, _ := observer.ObserveAddress("test")

	var receivedTransactions []Transaction
	for transaction := range transactionsChan {
		receivedTransactions = append(receivedTransactions, transaction)
	}

	if !reflect.DeepEqual(receivedTransactions, expectedTransactions) {
		t.Errorf("expected: %v, got: %v", expectedTransactions, receivedTransactions)
	}
}
