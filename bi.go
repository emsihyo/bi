package bi

import (
	"errors"
	"unsafe"
)

var (
	//ErrUnsupportedMethod ErrUnsupportedMethod
	ErrUnsupportedMethod = errors.New("bi unsupport method")
)

//BI BI
type BI struct {
	callers map[string]*caller
}

//NewBI NewBI
func NewBI() *BI {
	return &BI{callers: map[string]*caller{}}
}

//On On
func (bi *BI) On(m string, f interface{}) {
	bi.callers[m] = newCaller(f)
}

//Handle Handle
func (bi *BI) Handle(sess *Session) {
	sess.handle(bi)
}

func (bi *BI) onEvent(sessPtr unsafe.Pointer, method string, protocol Protocol, eventBytes []byte) (ackBytes []byte, err error) {
	caller, ok := bi.callers[method]
	if ok {
		return caller.call(sessPtr, protocol, eventBytes)
	}
	return nil, ErrUnsupportedMethod
}
