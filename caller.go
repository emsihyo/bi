package bi

import (
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

const (
	callerHelp = `
	e.g.
	func (*sess)
	func (*sess, *event)
	func (*sess, *event) *ack
	func (*sess) *ack
	func (*sess) *ack, Priority
	func (*sess, *event) *ack, Priority
	`
)

//caller caller
type caller struct {
	fun   reflect.Value
	tIn0  reflect.Type
	tIn1  reflect.Type
	tOut0 reflect.Type
	tOut1 reflect.Type
	newIn func() interface{}
	pool  sync.Pool
}

func newCaller(v interface{}) *caller {
	c := &caller{}
	vv := reflect.ValueOf(v)
	if vv.Kind() != reflect.Func {
		panic(errors.New(callerHelp))
	}
	c.fun = vv
	vt := vv.Type()
	switch vt.NumIn() {
	case 2:
		c.tIn1 = vt.In(1)
		if c.tIn1.Kind() != reflect.Ptr {
			panic(errors.New(callerHelp))
		}
		fallthrough
	case 1:
		c.tIn0 = vt.In(0)
		if c.tIn0.Kind() != reflect.Ptr {
			panic(errors.New(callerHelp))
		}
	default:
		panic(errors.New(callerHelp))
	}
	switch vt.NumOut() {
	case 2:
		c.tOut1 = vt.Out(1)
		if c.tOut1.Kind() != reflect.Int {
			panic(errors.New(callerHelp))
		}
		fallthrough
	case 1:
		c.tOut0 = vt.Out(0)
		if c.tOut0.Kind() != reflect.Ptr {
			panic(errors.New(callerHelp))
		}
	case 0:
	default:
		panic(errors.New(callerHelp))
	}
	c.pool.New = func() interface{} { return reflect.New(c.tIn1.Elem()).Interface() }
	return c
}

func (c *caller) call(sessPtr unsafe.Pointer, p Protocol, a []byte) ([]byte, Priority, error) {
	var vs []reflect.Value
	var err error
	sessValue := reflect.NewAt(c.tIn0.Elem(), sessPtr)
	if nil == c.tIn1 {
		vs = c.fun.Call([]reflect.Value{sessValue})
	} else {
		in1 := c.pool.Get()
		if err = p.Unmarshal(a, in1); nil != err {
			return nil, Normal, err
		}
		vs = c.fun.Call([]reflect.Value{sessValue, reflect.ValueOf(in1)})
		c.pool.Put(in1)
	}
	switch len(vs) {
	case 1:
		var b []byte
		if b, err = p.Marshal(vs[0].Interface()); nil != err {
			return nil, Normal, err
		}
		return b, Normal, nil
	case 2:
		var b []byte
		if b, err = p.Marshal(vs[0].Interface()); nil != err {
			return nil, Normal, err
		}
		return b, Priority(vs[1].Int()), nil
	default:
		return nil, Normal, nil
	}
}
