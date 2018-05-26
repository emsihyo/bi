package bi

import (
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//Pool Pool
var Pool = newPool()

type pool struct {
	TCPConn   *poolTCPConn
	WebSocket *poolWebSocketConn
	Timer     *poolTimer
	Payload   *poolPayload
}

func newPool() *pool {
	return &pool{TCPConn: newPoolTCPConn(), WebSocket: newPoolWebSocketConn(), Timer: newPoolTimer(), Payload: newPoolPayload()}
}

type poolTCPConn struct {
	sync.Pool
}

func newPoolTCPConn() *poolTCPConn {
	return &poolTCPConn{Pool: sync.Pool{New: func() interface{} { return &TCPConn{} }}}
}

//Get Get
func (p *poolTCPConn) Get(conn *net.TCPConn) *TCPConn {
	v := p.Pool.Get().(*TCPConn)
	v.conn = conn
	return v
}

//Put Put
func (p *poolTCPConn) Put(v *TCPConn) {
	v.conn = nil
	p.Pool.Put(v)
}

type poolWebSocketConn struct {
	sync.Pool
}

func newPoolWebSocketConn() *poolWebSocketConn {
	return &poolWebSocketConn{Pool: sync.Pool{New: func() interface{} { return &WebsocketConn{} }}}
}

//Get Get
func (p *poolWebSocketConn) Get(conn *websocket.Conn) *WebsocketConn {
	v := p.Pool.Get().(*WebsocketConn)
	v.conn = conn
	return v
}

//Put Put
func (p *poolWebSocketConn) Put(v *WebsocketConn) {
	v.conn = nil
	p.Pool.Put(v)
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
