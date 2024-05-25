package ethereum

import (
	"io"
	"log"
	"net/http"
	"sync"
)

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

// TransactionSerializer must be implemented by the struct
// that can be used to serialize transaction information
type TransactionSerializer interface {
	// SerializeTransaction serializes given transaction. It should be multiple goroutines safe.
	SerializeTransaction(transaction SerializableTransaction) error
}

// Transaction represents transaction from
// Ethereum.
type Transaction struct {
	ID      int64
	JsonRPC string
	Result  string
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
	observer   Observer
	serializer TransactionSerializer
	logger     *log.Logger
	httpClient *http.Client

	closeChan     chan struct{}
	subscribersWG sync.WaitGroup
}

var _ Parser = (*JSONRPCParser)(nil)
var _ io.Closer = (*JSONRPCParser)(nil)

// NewJSONRPCParser creates a new instance of JSONRPCParser.
func NewJSONRPCParser(
	observer Observer,
	serializer TransactionSerializer,
	httpClient *http.Client,
	logger *log.Logger,
) *JSONRPCParser {
	closeChan := make(chan struct{}, 1)

	return &JSONRPCParser{
		observer:   observer,
		httpClient: httpClient,
		logger:     logger,
		serializer: serializer,
		closeChan:  closeChan,
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

	return int(res)
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
	//TODO implement me
	panic("implement me")
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

			if err := jp.serializer.SerializeTransaction(SerializableTransaction{
				Address:     address,
				Transaction: transaction,
			}); err != nil {
				jp.logger.Printf("serialize transaction error: %s", err.Error())
			}
		}
	}
}
