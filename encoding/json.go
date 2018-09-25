package encoding

import (
	"bytes"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/encoding"
)

const (
	// Name is the name registered for the json encoder.
	Json = "json"
)

// codec implements encoding.Codec to encode messages into JSON.
type jsonCodec struct {
	marshaler   jsonpb.Marshaler
	unmarshaler jsonpb.Unmarshaler
}

func (c *jsonCodec) Marshal(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("not a proto message but %T: %v", v, v)
	}

	fmt.Printf("%v", msg)

	var w bytes.Buffer
	if err := c.marshaler.Marshal(&w, msg); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (c *jsonCodec) Unmarshal(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("not a proto message but %T: %v", v, v)
	}

	fmt.Printf("%v", msg)

	return c.unmarshaler.Unmarshal(bytes.NewReader(data), msg)
}

// Name returns the identifier of the codec.
func (c *jsonCodec) Name() string {
	return Json
}

func init() {
	encoding.RegisterCodec(new(jsonCodec))
}
