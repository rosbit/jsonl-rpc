package jsonlrpc

import (
	"encoding/json"
)

type JSONLResponse struct {
	Code       int `json:"code"`
	Msg     string `json:"msg"`
	HasJSONLs bool `json:"has-jsonls"`
}

type FnCollectJSONLs func(res json.RawMessage, jsonls <-chan interface{})
type JSONLClientResponse struct {
	JSONLResponse
	Result json.RawMessage `json:"result"`
	CollectJSONLs FnCollectJSONLs `json:"-"`
}

type JSONLServerResponse struct {
	JSONLResponse
	Result interface{} `json:"result"`
	JSONLs <-chan interface{} `json:"-"`
}
