package ethereum

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"tw/internal/clogger"
)

func TestJSONRPCParser_GetCurrentBlock(t *testing.T) {
	type fields struct {
		observer            Observer
		transactionsStorage TransactionsStorage
		logger              *log.Logger
		httpClient          *http.Client
		closeChan           chan struct{}
		subscribersWG       sync.WaitGroup
	}
	tests := []struct {
		name                      string
		mutateGetCurrentBlockFunc func()
		fields                    fields
		want                      int
	}{
		{
			name: "get current block request fails, returns 0",
			fields: fields{
				logger: clogger.ConsoleLogger,
			},
			mutateGetCurrentBlockFunc: func() {
				getCurrentBlockFunc = func(_ *http.Client) (string, error) {
					return "", errors.New("error while doing request")
				}
			},
			want: 0,
		},
		{
			name: "get current block request returns block number, returns the number",
			mutateGetCurrentBlockFunc: func() {
				getCurrentBlockFunc = func(_ *http.Client) (string, error) {
					return strconv.Itoa(0x2A), nil
				}
			},
			fields: fields{
				logger: clogger.ConsoleLogger,
			},
			want: 42,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mutateGetCurrentBlockFunc()

			jp := &JSONRPCParser{
				observer:            tt.fields.observer,
				transactionsStorage: tt.fields.transactionsStorage,
				logger:              tt.fields.logger,
				httpClient:          tt.fields.httpClient,
				closeChan:           tt.fields.closeChan,
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
				logger: clogger.ConsoleLogger,
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
				logger: clogger.ConsoleLogger,
				observerFunc: func() Observer {
					return &mockObserver{func(address string) (<-chan Transaction, error) {
						transactionsChan := make(chan Transaction)

						var wg sync.WaitGroup

						wg.Add(1)
						go func() {
							defer wg.Done()

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
				logger:              clogger.ConsoleLogger,
				transactionsStorage: &mockTransactionStorage{transactions: transactions},
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
	// not used in this context
	panic("implement me")
}

func (m *mockTransactionStorage) GetTransactionsForAddress(address string) []Transaction {
	return m.transactions
}
