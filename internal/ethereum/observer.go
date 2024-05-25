package ethereum

import "net/http"

type JSONRpcBasedObserver struct {
	httpClient *http.Client
}

func NewJSONRpcBasedObserver(httpClient *http.Client) *JSONRpcBasedObserver {
	return &JSONRpcBasedObserver{
		httpClient: httpClient,
	}
}

func (j *JSONRpcBasedObserver) ObserveAddress(address string) (<-chan Transaction, error) {
	//TODO implement me
	panic("implement me")
}
