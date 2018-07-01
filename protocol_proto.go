package bi

import (
	"github.com/gogo/protobuf/proto"
)

//ProtobufProtocol ProtobufProtocol
type ProtobufProtocol struct {
}

//Marshal Marshal
func (p *ProtobufProtocol) Marshal(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, ErrMarshal
	}
	return proto.Marshal(m)
}

//Unmarshal Unmarshal
func (p *ProtobufProtocol) Unmarshal(d []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return ErrUnmarshal
	}
	return proto.Unmarshal(d, m)
}
