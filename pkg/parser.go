package pkg

import (
	"net/http"
	"net/url"

	"tw/internal/clogger"
	"tw/internal/ethereum"
	"tw/internal/memory"
)

const cloudflareEthApiEndpoint = "https://cloudflare-eth.com"

type Parser = ethereum.Parser
type Transaction = ethereum.Transaction

func NewDefaultParser() Parser {
	apiUrl, _ := url.Parse(cloudflareEthApiEndpoint)
	// Tbh. http client could be passed as parameter here as well, as probably
	// only one will be used, but this is kind of refactored and I din't have time
	// to add it here, sr 😅
	apiWrapper := ethereum.NewEthApiWrapper(apiUrl)

	observer := ethereum.NewJSONRpcBasedObserver(http.DefaultClient, clogger.ConsoleLogger, apiWrapper)
	storage := memory.NewMemoryTransactionStorage()

	return ethereum.NewJSONRPCParser(
		observer,
		storage,
		apiWrapper,
		http.DefaultClient,
		clogger.ConsoleLogger,
	)
}
