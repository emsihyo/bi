package bi

import (
	"net"
	"testing"
	"time"
)

type PingTest struct {
	At int64 `json:"at,omitempty"`
}

type PongTest struct {
	At int64 `json:"at,omitempty"`
}

type TestConn struct {
	t int
	p Protocol
}

func NewTestConn() *TestConn {
	return &TestConn{t: 0, p: &JSONProtocol{}}
}

func (conn *TestConn) Close() {

}

func (conn *TestConn) Write(b []byte) error {
	return nil
}

func (conn *TestConn) Read() ([]byte, error) {
	<-time.After(time.Second * 2)
	conn.t++
	if conn.t > 2 {
		return nil, net.ErrWriteToConnected
	}
	ping := &PingTest{At: time.Now().Unix()}
	pingBytes, _ := conn.p.Marshal(ping)
	payload := &Payload{I: uint64(conn.t), T: Type_Event, M: "ping", A: pingBytes}
	payloadBytes, _ := conn.p.Marshal(payload)
	return payloadBytes, nil

}

func (conn *TestConn) RemoteAddr() string {
	return "addr"
}
func Test_Session(t *testing.T) {
	b := NewImpl()
	b.On("ping", func(sess *SessionImpl, ping *PingTest) (*PongTest, Weight) {
		return &PongTest{At: time.Now().Unix()}, Lazy
	})
	for i := 0; i < 1; i++ {
		go func() {
			sess := NewSessionImpl(NewTestConn(), &JSONProtocol{}, time.Second*30)
			b.Handle(sess)
		}()
	}
	sess := NewSessionImpl(NewTestConn(), &JSONProtocol{}, time.Second*30)
	b.Handle(sess)
}
