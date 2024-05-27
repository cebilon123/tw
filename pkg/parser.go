package pkg

import (
	"net/http"

	"tw/internal/clogger"
	"tw/internal/ethereum"
	"tw/internal/memory"
)

type Parser = ethereum.Parser
type Transaction = ethereum.Transaction

func NewDefaultParser() Parser {
	observer := ethereum.NewJSONRpcBasedObserver(http.DefaultClient, clogger.ConsoleLogger)
	storage := memory.NewMemoryTransactionStorage()

	return ethereum.NewJSONRPCParser(observer, storage, http.DefaultClient, clogger.ConsoleLogger)
}
