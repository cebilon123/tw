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

	methodGetCurrentBlock      = "eth_blockNumber"
	methodGetBlockByNumber     = "eth_getBlockByNumber"
	methodGetTransactionByHash = "eth_getTransactionByHash"
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

type getBlockByNumberResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		BaseFeePerGas         string        `json:"baseFeePerGas"`
		BlobGasUsed           string        `json:"blobGasUsed"`
		Difficulty            string        `json:"difficulty"`
		ExcessBlobGas         string        `json:"excessBlobGas"`
		ExtraData             string        `json:"extraData"`
		GasLimit              string        `json:"gasLimit"`
		GasUsed               string        `json:"gasUsed"`
		Hash                  string        `json:"hash"`
		LogsBloom             string        `json:"logsBloom"`
		Miner                 string        `json:"miner"`
		MixHash               string        `json:"mixHash"`
		Nonce                 string        `json:"nonce"`
		Number                string        `json:"number"`
		ParentBeaconBlockRoot string        `json:"parentBeaconBlockRoot"`
		ParentHash            string        `json:"parentHash"`
		ReceiptsRoot          string        `json:"receiptsRoot"`
		Sha3Uncles            string        `json:"sha3Uncles"`
		Size                  string        `json:"size"`
		StateRoot             string        `json:"stateRoot"`
		Timestamp             string        `json:"timestamp"`
		TotalDifficulty       string        `json:"totalDifficulty"`
		Transactions          []string      `json:"transactions"`
		TransactionsRoot      string        `json:"transactionsRoot"`
		Uncles                []interface{} `json:"uncles"`
		Withdrawals           []struct {
			Index          string `json:"index"`
			ValidatorIndex string `json:"validatorIndex"`
			Address        string `json:"address"`
			Amount         string `json:"amount"`
		} `json:"withdrawals"`
		WithdrawalsRoot string `json:"withdrawalsRoot"`
	} `json:"result"`
}

func getTransactionsForBlock(httpClient *http.Client, blockNum string) ([]Transaction, error) {
	ethReq := ethRequest{
		ID:      generateRandomID(),
		JSONRpc: defaultJSONRpc,
		Method:  methodGetBlockByNumber,
		Params: []string{
			blockNum,
			"false",
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

	var ethRes getBlockByNumberResponse
	if err := json.Unmarshal(readBytes, &ethRes); err != nil {
		return nil, fmt.Errorf("json unmarshal bytes from response: %w", err)
	}

	var transactions []Transaction

	for _, transactionHash := range ethRes.Result.Transactions {
		transaction, err := getTransaction(httpClient, transactionHash)
		if err != nil {
			return nil, fmt.Errorf("fetching transactions: %w", err)
		}

		transactions = append(transactions, *transaction)
	}

	return transactions, nil
}

type getTransactionByHashResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  Transaction
}

func getTransaction(httpClient *http.Client, transactionHash string) (*Transaction, error) {
	ethReq := ethRequest{
		ID:      generateRandomID(),
		JSONRpc: defaultJSONRpc,
		Method:  methodGetTransactionByHash,
		Params: []string{
			transactionHash,
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

	var ethRes getTransactionByHashResponse
	if err := json.Unmarshal(readBytes, &ethRes); err != nil {
		return nil, fmt.Errorf("json unmarshal bytes from response: %w", err)
	}

	return &ethRes.Result, nil
}

func generateRandomID() int64 {
	return rand.Int64()
}
