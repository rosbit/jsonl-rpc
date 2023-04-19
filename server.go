package jsonlrpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/rpc"
)

type serverCodec struct {
	dec *json.Decoder // for reading JSON values
	enc *json.Encoder // for writing JSON values
	c   io.Closer
}

// NewServerCodec returns a new rpc.ServerCodec on conn.
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{
		dec:     createJSONDecoder(conn),
		enc:     createJSONEncoder(conn),
		c:       conn,
	}
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	var req rpc.Request
	if err := c.dec.Decode(&req); err != nil {
		return err
	}
	r.ServiceMethod = req.ServiceMethod
	r.Seq = req.Seq
	return nil
}

func (c *serverCodec) ReadRequestBody(x any) error {
	if x == nil {
		return nil
	}
	return c.dec.Decode(x)
}

func (c *serverCodec) WriteResponse(r *rpc.Response, x any) error {
	if err := c.enc.Encode(r); err != nil {
		return err
	}
	if x == nil {
		return fmt.Errorf("repsone expected")
	}
	resp, ok := x.(*JSONLServerResponse)
	if !ok {
		return fmt.Errorf("response body must be type of JSONLRepsonse")
	}
	if err := c.enc.Encode(resp); err != nil {
		return err
	}
	if resp.HasJSONLs {
		for row := range resp.JSONLs {
			if err := c.enc.Encode(row); err != nil {
				return err
			}
		}
		c.enc.Encode(nil)
	}
	return nil
}

func (c *serverCodec) Close() error {
	return c.c.Close()
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
func ServeConn(conn io.ReadWriteCloser) {
	rpc.ServeCodec(NewServerCodec(conn))
}
