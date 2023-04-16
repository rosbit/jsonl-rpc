package jsonlrpc

import (
	"encoding/json"
	"io"
)

func createJSONEncoder(conn io.ReadWriteCloser) (*json.Encoder) {
	enc := json.NewEncoder(conn)
	enc.SetEscapeHTML(false)
	return enc
}

func createJSONDecoder(conn io.ReadWriteCloser) (*json.Decoder) {
	return json.NewDecoder(conn)
}

