package bi

import (
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

//Pool Pool
var Pool = newPool()

type pool struct {
	TCPConn   *poolTCPConn
	WebSocket *poolWebSocketConn
	Timer     *poolTimer
	Session   *poolSession
	pkg       *poolPackage
}

func newPool() *pool {
	return &pool{TCPConn: newPoolTCPConn(), WebSocket: newPoolWebSocketConn(), Timer: newPoolTimer(), Session: newPoolSession(), pkg: newPoolPackage()}
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

type poolSession struct {
	sync.Pool
}

func newPoolSession() *poolSession {
	return &poolSession{Pool: sync.Pool{New: func() interface{} {
		return &Session{hand: newHandler(), didDisconnects: []chan error{}, didReceivePackage: make(chan *Package), didReceiveError: make(chan error, 1), marshalledEvent: make(chan []byte, 512)}
	}}}
}

//Get Get
func (p *poolSession) Get(conn Conn, protocol Protocol, timeout time.Duration) *Session {
	v := p.Pool.Get().(*Session)
	v.id = uuid.NewV3(uuid.NewV4(), conn.RemoteAddr()).String()
	v.conn = conn
	v.protocol = protocol
	v.timeout = timeout
	return v
}

//Put Put
func (p *poolSession) Put(v *Session) {
	v.didDisconnects = []chan error{}
	v.hand.reset()
	select {
	case <-v.timer.C:
	default:
	}
loop:
	for {
		select {
		case <-v.didReceiveError:
		case <-v.didReceivePackage:
		case <-v.marshalledEvent:
		default:
			break loop
		}
	}
	p.Pool.Put(v)
}

type poolPackage struct {
	sync.Pool
}

func newPoolPackage() *poolPackage {
	return &poolPackage{Pool: sync.Pool{New: func() interface{} { return &Package{} }}}
}

//Get Get
func (p *poolPackage) Get() *Package {
	v := p.Pool.Get().(*Package)
	return v
}

//Put Put
func (p *poolPackage) Put(v *Package) {
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
