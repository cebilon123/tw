package ethereum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
)

const (
	cloudflareEthApiEndpoint = "https://cloudflare-eth.com"
	defaultJSONRpc           = "2.0"

	methodGetCurrentBlock     = "eth_blockNumber"
	methodGetTransactionCount = "eth_getTransactionCount"
)

type ethRequest struct {
	ID      int64    `json:"id"`
	JSONRpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

type ethBaseResponse struct {
	JSONRpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

// getCurrentBlock returns current block number based on the http call
// to the api.
func getCurrentBlock(httpClient *http.Client) (string, error) {
	ethReq := ethRequest{
		ID:      generateRandomID(),
		JSONRpc: defaultJSONRpc,
		Method:  methodGetCurrentBlock,
		Params:  []string{},
	}

	body, err := json.Marshal(ethReq)
	if err != nil {
		return "", fmt.Errorf("json marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cloudflareEthApiEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("new http request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http client do request: %w", err)
	}

	if res.StatusCode < 200 && res.StatusCode > 299 {
		return "", fmt.Errorf("invalid response status code from api")
	}

	readBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read all bytes from response: %w", err)
	}

	var ethRes ethBaseResponse
	if err := json.Unmarshal(readBytes, &ethRes); err != nil {
		return "", fmt.Errorf("json unmarshal bytes from response: %w", err)
	}

	return ethRes.Result, nil
}

func getTransactionsForBlock(httpClient *http.Client, blockNum string) ([]Transaction, error) {
	ethReq := ethRequest{
		ID:      generateRandomID(),
		JSONRpc: defaultJSONRpc,
		Method:  methodGetCurrentBlock,
		Params: []string{
			blockNum,
		},
	}

	body, err := json.Marshal(ethReq)
	if err != nil {
		return nil, fmt.Errorf("json marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cloudflareEthApiEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("new http request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http client do request: %w", err)
	}

	if res.StatusCode < 200 && res.StatusCode > 299 {
		return nil, fmt.Errorf("invalid response status code from api")
	}

	readBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read all bytes from response: %w", err)
	}

	var ethRes ethBaseResponse
	if err := json.Unmarshal(readBytes, &ethRes); err != nil {
		return nil, fmt.Errorf("json unmarshal bytes from response: %w", err)
	}

	return n.Int64(), nil
}

func generateRandomID() int64 {
	return rand.Int64()
}
