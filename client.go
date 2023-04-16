package jsonlrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/rpc"
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
	rows := make(chan interface{})
	go resp.CollectJSONLs(rows)
	for {
		var row interface{}
		if err := c.dec.Decode(&row); err != nil {
			break
		}
		if row == nil {
			break
		}
		rows <- row
	}
	close(rows)

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
	return NewClient(conn), err
}