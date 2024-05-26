package pkg

import (
	"log"
	"net/http"
	"tw/internal/ethereum"
	"tw/internal/memory"
)

type Parser = ethereum.Parser
type Transaction = ethereum.Transaction

func NewDefaultParser() Parser {
	cmdLogger := log.New(&consoleWriter{}, "PARSER: ", log.Ldate|log.Ltime|log.Lshortfile)
	observer := ethereum.NewJSONRpcBasedObserver(http.DefaultClient, cmdLogger)
	storage := memory.NewMemoryTransactionStorage()

	return ethereum.NewJSONRPCParser(observer, storage, http.DefaultClient, cmdLogger)
}

type consoleWriter struct {
}

func (c *consoleWriter) Write(p []byte) (n int, err error) {
	log.Println(string(p))

	return len(p), nil
}
