package ethereum

import (
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
)

// maxBlockDepth describes how many blocks from the current one
// we are going to look for the transactions.
const maxBlockDepth = 20

// Parser must be implemented by any struct
// that can communicate with the ethereum
// in order to work with the blockchain.
type Parser interface {
	// GetCurrentBlock returns last parsed block
	GetCurrentBlock() int
	// Subscribe adds address to observer
	Subscribe(address string) bool
	// GetTransactions lists inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

// Observer must be implemented by any struct
// used to observe changes on blockchain.
type Observer interface {
	// ObserveAddress observes given address and returns transaction channel that can be watched for incoming
	// transactions.
	ObserveAddress(address string) (<-chan Transaction, error)
}

// TransactionsStorage should be implemented
// by the struct that can store transactions.
type TransactionsStorage interface {
	// SerializeTransaction serializes given transaction. It should be multiple goroutines safe.
	SerializeTransaction(transaction SerializableTransaction) error
	// GetTransactionsForAddress returns transactions for a given
	// address. It should be multiple goroutines safe.
	GetTransactionsForAddress(address string) []Transaction
}

// Transaction represents transaction from
// Ethereum.
type Transaction struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Type                 string `json:"type"`
		BlockHash            string `json:"blockHash"`
		BlockNumber          string `json:"blockNumber"`
		From                 string `json:"from"`
		Gas                  string `json:"gas"`
		Hash                 string `json:"hash"`
		Input                string `json:"input"`
		Nonce                string `json:"nonce"`
		To                   string `json:"to"`
		TransactionIndex     string `json:"transactionIndex"`
		Value                string `json:"value"`
		V                    string `json:"v"`
		R                    string `json:"r"`
		S                    string `json:"s"`
		GasPrice             string `json:"gasPrice"`
		MaxFeePerGas         string `json:"maxFeePerGas"`
		MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
		ChainId              string `json:"chainId"`
		AccessList           []struct {
			Address     string   `json:"address"`
			StorageKeys []string `json:"storageKeys"`
		} `json:"accessList"`
	} `json:"result"`
}

// SerializableTransaction represents transaction
// that can be serialized. The only additional field
// is the address.
type SerializableTransaction struct {
	Address     string `json:"address"`
	Transaction `json:"transaction"`
}

// JSONRPCParser used to parse transactions
// with the help of JSONRPC.
type JSONRPCParser struct {
	observer            Observer
	transactionsStorage TransactionsStorage
	logger              *log.Logger
	httpClient          *http.Client

	closeChan     chan struct{}
	subscribersWG sync.WaitGroup
}

var _ Parser = (*JSONRPCParser)(nil)
var _ io.Closer = (*JSONRPCParser)(nil)

// NewJSONRPCParser creates a new instance of JSONRPCParser.
func NewJSONRPCParser(
	observer Observer,
	transactionsStorage TransactionsStorage,
	httpClient *http.Client,
	logger *log.Logger,
) *JSONRPCParser {
	closeChan := make(chan struct{}, 1)

	return &JSONRPCParser{
		observer:            observer,
		httpClient:          httpClient,
		logger:              logger,
		transactionsStorage: transactionsStorage,
		closeChan:           closeChan,
	}
}

// Close safely closes JSONRPCParser. It sends close signal to
// serialize goroutine, which then safely is being done
// with the serialize process.
func (jp *JSONRPCParser) Close() error {
	jp.closeChan <- struct{}{}

	// wait for the serialization to be done for all addresses
	jp.subscribersWG.Wait()

	return nil
}

// GetCurrentBlock returns an number of current block.
func (jp *JSONRPCParser) GetCurrentBlock() int {
	res, err := getCurrentBlock(jp.httpClient)
	if err != nil {
		jp.logger.Printf("get current block error: %s", err.Error())
		return 0
	}

	n := new(big.Int)
	// passing 0, it will pick base based on the string
	n.SetString(res, 0)

	return int(n.Int64())
}

func (jp *JSONRPCParser) Subscribe(address string) bool {
	transactionsChan, err := jp.observer.ObserveAddress(address)
	if err != nil {
		jp.logger.Printf("observer observe address: %s", err.Error())
		return false
	}

	jp.subscribersWG.Add(1)

	go jp.onTransactionsSubscribe(address, transactionsChan)

	return true
}

func (jp *JSONRPCParser) GetTransactions(address string) []Transaction {
	return jp.transactionsStorage.GetTransactionsForAddress(address)
}

func (jp *JSONRPCParser) onTransactionsSubscribe(address string, transactionsChan <-chan Transaction) {
	defer func() {
		jp.logger.Printf("on transaction subscribe done for address: %s", address)
		jp.subscribersWG.Done()
	}()

	for {
		select {
		case <-jp.closeChan:
			jp.logger.Println("signal from close chan")
			return
		case transaction, ok := <-transactionsChan:
			if !ok {
				jp.logger.Println("transaction chan closed")
				return
			}

			if err := jp.transactionsStorage.SerializeTransaction(SerializableTransaction{
				Address:     address,
				Transaction: transaction,
			}); err != nil {
				jp.logger.Printf("serialize transaction error: %s", err.Error())
			}
		}
	}
}
