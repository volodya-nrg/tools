package rpcbus

type request struct {
	JsonRpc string `json:"jsonrpc,omitempty"`
	ID      string `json:"id,omitempty"` // id запрос-ответ
	Method  string `json:"method,omitempty"`
	Params  any    `json:"params,omitempty"`
}

type Response struct {
	JsonRpc string `json:"jsonrpc"`
	ID      string `json:"id"` // id запрос-ответ
	Result  any    `json:"result"`
}
