package memory

import (
	"sync"

	"tw/internal/ethereum"
)

// TransactionSerializer is simple in memory serializer
// for the transactions.
type TransactionSerializer struct {
	transactions []ethereum.SerializableTransaction

	mu sync.Mutex
}

var _ ethereum.TransactionSerializer = (*TransactionSerializer)(nil)

func NewMemoryTransactionSerializer() *TransactionSerializer {
	return &TransactionSerializer{}
}

// SerializeTransaction serializes transactions to the memory.
func (ts *TransactionSerializer) SerializeTransaction(transaction ethereum.SerializableTransaction) error {
	// it could be used by multiple clients simultaneously
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.transactions = append(ts.transactions, transaction)

	return nil
}
