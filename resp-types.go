package jsonlrpc

import (
	"encoding/json"
)

type JSONLResponse struct {
	Code       int `json:"code"`
	Msg     string `json:"msg"`
	HasJSONLs bool `json:"has-jsonls"`
}

type JSONLClientResponse struct {
	JSONLResponse
	Result json.RawMessage `json:"result"`
	CollectJSONLs func(jsonls <-chan interface{}) `json:"-"`
}

type JSONLServerResponse struct {
	JSONLResponse
	Result interface{} `json:"result"`
	JSONLs <-chan interface{} `json:"-"`
}
