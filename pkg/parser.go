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
	observer := ethereum.NewJSONRpcBasedObserver(http.DefaultClient)
	serializer := memory.NewMemoryTransactionSerializer()
	cmdLogger := log.New(os.Stdout, "PARSER: ", log.Ldate|log.Ltime|log.Lshortfile)

	return ethereum.NewJSONRPCParser(observer, serializer, http.DefaultClient, cmdLogger)
}
