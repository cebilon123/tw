package memory

import (
	"sync"

	"tw/internal/ethereum"
)

// TransactionMemoryStorage is simple in memory storage
// for the transactions.
type TransactionMemoryStorage struct {
	transactionsMap map[string][]ethereum.Transaction

	mu   sync.Mutex
	rwMu sync.RWMutex
}

var _ ethereum.TransactionsStorage = (*TransactionMemoryStorage)(nil)

func NewMemoryTransactionStorage() *TransactionMemoryStorage {
	return &TransactionMemoryStorage{
		transactionsMap: make(map[string][]ethereum.Transaction),
	}
}

// SerializeTransaction serializes transactions to the memory.
func (ts *TransactionMemoryStorage) SerializeTransaction(serializableTransaction ethereum.SerializableTransaction) error {
	// it could be used by multiple clients simultaneously
	ts.mu.Lock()
	defer ts.mu.Unlock()

	_, ok := ts.transactionsMap[serializableTransaction.Address]
	if !ok {
		ts.transactionsMap[serializableTransaction.Address] = []ethereum.Transaction{serializableTransaction.Transaction}

		return nil
	}

	ts.transactionsMap[serializableTransaction.Address] = append(ts.transactionsMap[serializableTransaction.Address], serializableTransaction.Transaction)

	return nil
}

func (ts *TransactionMemoryStorage) GetTransactionsForAddress(address string) []ethereum.Transaction {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	v, ok := ts.transactionsMap[address]
	if !ok {
		return nil
	}

	return v
}
