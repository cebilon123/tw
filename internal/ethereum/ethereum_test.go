package ethereum

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"tw/internal/clogger"
)

func TestJSONRPCParser_GetCurrentBlock(t *testing.T) {
	type fields struct {
		observer            Observer
		transactionsStorage TransactionsStorage
		apiWrapper          *mockApiWrapper
		logger              *log.Logger
		httpClient          *http.Client
		closeChan           chan struct{}
		subscribersWG       sync.WaitGroup
	}
	tests := []struct {
		name                 string
		mutateApiWrapperFunc func(apiWrapper *mockApiWrapper)
		fields               fields
		want                 int
	}{
		{
			name: "get current block request fails, returns 0",
			fields: fields{
				logger:              clogger.ConsoleLogger,
				apiWrapper:          &mockApiWrapper{},
				transactionsStorage: &mockTransactionStorage{},
			},
			mutateApiWrapperFunc: func(apiWrapper *mockApiWrapper) {
				apiWrapper.getCurrentBlockFunc = func(httpClient *http.Client) (string, error) {
					return "", errors.New("error")
				}
			},
			want: 0,
		},
		{
			name: "get current block request returns block number, returns the number",
			mutateApiWrapperFunc: func(apiWrapper *mockApiWrapper) {
				apiWrapper.getCurrentBlockFunc = func(httpClient *http.Client) (string, error) {
					return "0x2A", nil
				}
			},
			fields: fields{
				logger:              clogger.ConsoleLogger,
				apiWrapper:          &mockApiWrapper{},
				transactionsStorage: &mockTransactionStorage{},
			},
			want: 42,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mutateApiWrapperFunc(tt.fields.apiWrapper)

			jp := &JSONRPCParser{
				observer:            tt.fields.observer,
				transactionsStorage: tt.fields.transactionsStorage,
				logger:              tt.fields.logger,
				httpClient:          tt.fields.httpClient,
				closeChan:           tt.fields.closeChan,
				apiWrapper:          tt.fields.apiWrapper,
				subscribersWG:       tt.fields.subscribersWG,
			}
			if got := jp.GetCurrentBlock(); got != tt.want {
				t.Errorf("GetCurrentBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONRPCParser_Subscribe(t *testing.T) {
	type fields struct {
		observerFunc        func() Observer
		transactionsStorage TransactionsStorage
		logger              *log.Logger
		httpClient          *http.Client
		closeChan           chan struct{}
		subscribersWG       sync.WaitGroup
	}
	type args struct {
		address string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "observer returns error, should return false",
			fields: fields{
				logger:              clogger.ConsoleLogger,
				transactionsStorage: &mockTransactionStorage{},
				observerFunc: func() Observer {
					return &mockObserver{
						func(address string) (<-chan Transaction, error) {
							return nil, fmt.Errorf("error returned")
						},
					}
				},
			},
			args: args{
				address: "test",
			},
			want: false,
		},
		{
			name: "observer returns channel, should return true",
			fields: fields{
				logger:              clogger.ConsoleLogger,
				transactionsStorage: &mockTransactionStorage{},
				observerFunc: func() Observer {
					return &mockObserver{func(address string) (<-chan Transaction, error) {
						transactionsChan := make(chan Transaction)

						var wg sync.WaitGroup

						wg.Add(1)
						go func() {
							defer wg.Done()

							time.Sleep(time.Second * 5)

							for range 3 {
								transactionsChan <- Transaction{
									From: address,
								}
							}
						}()

						go func() {
							wg.Wait()

							close(transactionsChan)
						}()

						return transactionsChan, nil
					}}
				},
			},
			args: args{
				address: "test",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jp := &JSONRPCParser{
				observer:            tt.fields.observerFunc(),
				transactionsStorage: tt.fields.transactionsStorage,
				logger:              tt.fields.logger,
				httpClient:          tt.fields.httpClient,
				closeChan:           tt.fields.closeChan,
				subscribersWG:       tt.fields.subscribersWG,
			}
			if got := jp.Subscribe(tt.args.address); got != tt.want {
				t.Errorf("Subscribe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONRPCParser_GetTransactions(t *testing.T) {
	transactions := []Transaction{
		{
			From: "test",
		},
	}

	type fields struct {
		observer            Observer
		transactionsStorage TransactionsStorage
		logger              *log.Logger
		httpClient          *http.Client
		closeChan           chan struct{}
		subscribersWG       sync.WaitGroup
	}
	type args struct {
		address string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Transaction
	}{
		{
			name: "there are no transactions in storage, returns empty array",
			fields: fields{
				transactionsStorage: &mockTransactionStorage{},
				logger:              clogger.ConsoleLogger,
			},
			args: args{
				address: "test",
			},
			want: nil,
		},
		{
			name: "transaction storage returns transactions, returns these transactions",
			fields: fields{
				transactionsStorage: &mockTransactionStorage{transactions: transactions},
				logger:              clogger.ConsoleLogger,
			},
			args: args{},
			want: transactions,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jp := &JSONRPCParser{
				observer:            tt.fields.observer,
				transactionsStorage: tt.fields.transactionsStorage,
				logger:              tt.fields.logger,
				httpClient:          tt.fields.httpClient,
				closeChan:           tt.fields.closeChan,
				subscribersWG:       tt.fields.subscribersWG,
			}
			if got := jp.GetTransactions(tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTransactions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockObserver struct {
	observeAddressFunc func(address string) (<-chan Transaction, error)
}

func (m *mockObserver) ObserveAddress(address string) (<-chan Transaction, error) {
	if m.observeAddressFunc != nil {
		return m.observeAddressFunc(address)
	}

	return nil, nil
}

type mockTransactionStorage struct {
	transactions []Transaction
}

func (m *mockTransactionStorage) SerializeTransaction(transaction SerializableTransaction) error {
	m.transactions = append(m.transactions, transaction.Transaction)

	return nil
}

func (m *mockTransactionStorage) GetTransactionsForAddress(address string) []Transaction {
	return m.transactions
}

type mockApiWrapper struct {
	getCurrentBlockFunc         func(httpClient *http.Client) (string, error)
	getTransactionsForBlockFunc func(httpClient *http.Client, blockNum string) ([]Transaction, error)
}

func (m *mockApiWrapper) GetCurrentBlock(httpClient *http.Client) (string, error) {
	if m.getCurrentBlockFunc != nil {
		return m.getCurrentBlockFunc(httpClient)
	}

	return "", nil
}

func (m *mockApiWrapper) GetTransactionsForBlock(httpClient *http.Client, blockNum string) ([]Transaction, error) {
	if m.getTransactionsForBlockFunc != nil {
		return m.getTransactionsForBlockFunc(httpClient, blockNum)
	}

	return nil, nil
}
