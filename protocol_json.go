package bi

import (
	"encoding/json"
)

//JSONProtocol JSONProtocol
type JSONProtocol struct {
}

//Marshal Marshal
func (p *JSONProtocol) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

//Unmarshal Unmarshal
func (p *JSONProtocol) Unmarshal(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}
