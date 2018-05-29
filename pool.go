package bi

import (
	"sync"
	"time"
)

//Pool Pool
var Pool = newPool()

type pool struct {
	Timer   *poolTimer
	Payload *poolPayload
}

func newPool() *pool {
	return &pool{Timer: newPoolTimer(), Payload: newPoolPayload()}
}

type poolPayload struct {
	sync.Pool
}

func newPoolPayload() *poolPayload {
	return &poolPayload{Pool: sync.Pool{New: func() interface{} { return &Payload{} }}}
}

//Get Get
func (p *poolPayload) Get() *Payload {
	v := p.Pool.Get().(*Payload)
	return v
}

//Put Put
func (p *poolPayload) Put(v *Payload) {
	v.I = 0
	v.M = ""
	v.A = []byte{}
	v.T = Type_Event
	p.Pool.Put(v)
}

type poolTimer struct {
	sync.Pool
}

func newPoolTimer() *poolTimer {
	return &poolTimer{Pool: sync.Pool{}}
}

//Get Get
func (p *poolTimer) Get(timeout time.Duration) *time.Timer {
	if v, _ := p.Pool.Get().(*time.Timer); v != nil {
		v.Reset(timeout)
		return v
	}
	return time.NewTimer(timeout)
}

//Put Put
func (p *poolTimer) Put(v *time.Timer) {
	if !v.Stop() {
		select {
		case <-v.C:
		default:
		}
	}
	p.Pool.Put(v)
}
