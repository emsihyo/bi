package bi

import (
	"net"
	"testing"
	"time"
)

type pingTest struct {
	At int64 `json:"at,omitempty"`
}

type pongTest struct {
	At int64 `json:"at,omitempty"`
}

type connTest struct {
	t int
	p Protocol
}

func newConnTest() *connTest {
	return &connTest{t: 0, p: &JSONProtocol{}}
}

func (conn *connTest) Close() {

}

func (conn *connTest) Write(b []byte) error {
	return nil
}

func (conn *connTest) Read() ([]byte, error) {
	<-time.After(time.Millisecond * 100)
	conn.t++
	if conn.t > 10 {
		return nil, net.ErrWriteToConnected
	}
	ping := &pingTest{At: time.Now().UnixNano()}
	pingBytes, _ := conn.p.Marshal(ping)
	payload := &Payload{I: uint64(conn.t), T: Type_Event, M: "ping", A: pingBytes}
	payloadBytes, _ := conn.p.Marshal(payload)
	return payloadBytes, nil

}

func (conn *connTest) RemoteAddr() string {
	return "addr"
}

func Test_Session(t *testing.T) {
	b := NewImpl()
	b.On(Connection, func(sess *SessionImpl) {
		t.Log("connection")
	})
	b.On(Disconnection, func(sess *SessionImpl) {
		t.Log("disconnection")
	})
	b.On("ping", func(sess *SessionImpl, ping *pingTest) (*pongTest, Weight) {
		t.Log("event:", *ping)
		return &pongTest{At: time.Now().UnixNano()}, Lazy
	})
	sess := NewSessionImpl(newConnTest(), &JSONProtocol{}, time.Second*30)
	b.Handle(sess)
}
