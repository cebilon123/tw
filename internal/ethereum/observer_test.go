package ethereum

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

func TestJSONRpcBasedObserver_ObserveAddress_Then_GetTransactions(t *testing.T) {
	var mu sync.Mutex

	currBlockNum := 0

	getCurrentBlockFunc = func(httpClient *http.Client) (string, error) {
		mu.Lock()
		defer mu.Unlock()

		currBlockNum++

		return fmt.Sprintf("%x", currBlockNum), nil
	}

	var expectedTransactions []Transaction

	transactionNum := 0

	getTransactionsForBlockFunc = func(httpClient *http.Client, blockNum string) ([]Transaction, error) {
		mu.Lock()
		defer mu.Unlock()

		transactions := []Transaction{
			{
				From: "test",
				To:   fmt.Sprintf("to%d", transactionNum+1),
			},
		}

		expectedTransactions = append(expectedTransactions, transactions...)

		transactionNum++

		return transactions, nil
	}

	observer := JSONRpcBasedObserver{
		httpClient: http.DefaultClient,
		closeChan:  make(chan struct{}, 1),
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
