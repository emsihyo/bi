package bi

import (
	"unsafe"
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

func (bi *BI) onRequest(sessPtr unsafe.Pointer, m string, p Protocol, a []byte) ([]byte, error) {
	caller, ok := bi.callers[m]
	if ok {
		return caller.call(sessPtr, p, a)
	}
	return nil, nil
}
