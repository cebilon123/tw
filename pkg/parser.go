package pkg

import (
	"log"
	"net/http"
	"os"

	"tw/internal/ethereum"
	"tw/internal/memory"
)

type Parser = ethereum.Parser

func NewDefaultParser() Parser {
	cmdLogger := log.New(os.Stdout, "PARSER: ", log.Ldate|log.Ltime|log.Lshortfile)
	observer := ethereum.NewJSONRpcBasedObserver(http.DefaultClient, cmdLogger)
	storage := memory.NewMemoryTransactionStorage()

	return ethereum.NewJSONRPCParser(observer, storage, http.DefaultClient, cmdLogger)
}
