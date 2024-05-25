package ethereum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
)

const (
	cloudflareEthApiEndpoint = "https://cloudflare-eth.com"
	defaultJSONRpc           = "2.0"

	methodGetCurrentBlock = "eth_subscribe"
)

type ethRequest struct {
	ID      int      `json:"id"`
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
func getCurrentBlock(httpClient *http.Client) (int64, error) {
	ethReq := ethRequest{
		JSONRpc: defaultJSONRpc,
		Method:  methodGetCurrentBlock,
		Params:  []string{},
	}

	body, err := json.Marshal(ethReq)
	if err != nil {
		return 0, fmt.Errorf("json marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, cloudflareEthApiEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return 0, fmt.Errorf("new http request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http client do request: %w", err)
	}

	if res.StatusCode < 200 && res.StatusCode > 299 {
		return 0, fmt.Errorf("invalid response status code from api")
	}

	readBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("read all bytes from response: %w", err)
	}

	var ethRes ethBaseResponse
	if err := json.Unmarshal(readBytes, &ethRes); err != nil {
		return 0, fmt.Errorf("json unmarshal bytes from response: %w", err)
	}

	n := new(big.Int)
	// passing 0, it will pick base based on the string
	n.SetString(ethRes.Result, 0)

	return n.Int64(), nil
}
