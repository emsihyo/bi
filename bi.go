package bi

import (
	"errors"
	"unsafe"
)

var (
	//ErrUnsupportedMethod ErrUnsupportedMethod
	ErrUnsupportedMethod = errors.New("bi unsupport method")
	//ErrTooLargePayload ErrTooLargePayload
	ErrTooLargePayload = errors.New("bi.session pay load is too large")
	//ErrChanFull ErrChanFull
	ErrChanFull = errors.New("bi.session chan is full")
	//ErrTimeOut ErrTimeOut
	ErrTimeOut = errors.New("bi.session timeout")
	//ErrClosed ErrClosed
	ErrClosed = errors.New("bi.session closed")
)

//BI BI
type BI interface {
	OnEvent(sessPtr unsafe.Pointer, method string, protocol Protocol, eventBytes []byte) (ackBytes []byte, weight Weight, err error)
}

//Impl Impl
type Impl struct {
	callers map[string]*caller
}

//NewImpl NewImpl
func NewImpl() *Impl {
	return &Impl{callers: map[string]*caller{}}
}

//On On
func (impl *Impl) On(m string, f interface{}) {
	impl.callers[m] = newCaller(f)
}

//Handle Handle
func (impl *Impl) Handle(sess Session) {
	sess.handle(impl)
}

//OnEvent OnEvent
func (impl *Impl) OnEvent(sessPtr unsafe.Pointer, method string, protocol Protocol, eventBytes []byte) (ackBytes []byte, weight Weight, err error) {
	caller, ok := impl.callers[method]
	if ok {
		return caller.call(sessPtr, protocol, eventBytes)
	}
	return nil, Normal, ErrUnsupportedMethod
}
