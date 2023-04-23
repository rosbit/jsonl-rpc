package jsonlrpc

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	"time"
	"fmt"
	"io"
	"os"
	"net"
	"net/rpc"
	"net/http"
)

type clientCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer
}

// NewClientCodec returns a new rpc.ClientCodec on conn.
func NewClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &clientCodec{
		dec:     createJSONDecoder(conn),
		enc:     createJSONEncoder(conn),
		c:       conn,
	}
}

func (c *clientCodec) WriteRequest(r *rpc.Request, param any) error {
	if err := c.enc.Encode(r); err != nil {
		return err
	}
	return c.enc.Encode(param)
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	var resp rpc.Response
	if err := c.dec.Decode(&resp); err != nil {
		return err
	}
	r.ServiceMethod = resp.ServiceMethod
	r.Seq = resp.Seq
	r.Error = resp.Error
	return nil
}

func (c *clientCodec) ReadResponseBody(x any) error {
	if x == nil {
		return fmt.Errorf("response body expected\n")
	}
	resp, ok := x.(*JSONLClientResponse)
	if !ok {
		return fmt.Errorf("response body must be type of JSONLClientResponse")
	}
	if err := c.dec.Decode(&resp); err != nil {
		return err
	}
	if !resp.HasJSONLs {
		return nil
	}

	jsonls := make(chan interface{})
	go resp.CollectJSONLs(resp.Result, jsonls) // return the result in advance

	// all the JSONL must be be read in the current goruntine
	for {
		var jsonl interface{}
		if err := c.dec.Decode(&jsonl); err != nil {
			break
		}
		if jsonl == nil {
			break
		}
		jsonls <- jsonl
	}
	close(jsonls)

	return nil
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}

// NewClient returns a new rpc.Client to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewClientCodec(conn))
}

// Dial connects to server at the specified network address.
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	websocketEndpoint := os.Getenv("WEBSOCKET_ENDPOINT")
	if len(websocketEndpoint) > 0 {
		url := fmt.Sprintf("ws://%s%s", address, websocketEndpoint)
		wsDialer := &websocket.Dialer{
			Proxy: http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			NetDial: func(_, _ string) (net.Conn, error) {
				return conn, nil
			},
		}
		if _, _, err = wsDialer.Dial(url, nil); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return NewClient(conn), err
}
