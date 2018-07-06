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
	i        int
	protocol Protocol
}

func newConnTest() *connTest {
	return &connTest{i: 0, protocol: &JSONProtocol{}}
}

func (conn *connTest) Close() {

}

func (conn *connTest) Write(b []byte) error {
	return nil
}

func (conn *connTest) Read() ([]byte, error) {
	<-time.After(time.Millisecond)
	conn.i++
	if conn.i > 10000 {
		return nil, net.ErrWriteToConnected
	}
	ping := &pingTest{At: time.Now().UnixNano()}
	pingBytes, _ := conn.protocol.Marshal(ping)
	payload := &Payload{I: uint32(conn.i), T: Type_Event, M: "ping", A: pingBytes}
	payloadBytes, _ := conn.protocol.Marshal(payload)
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
	b.On("ping", func(sess *SessionImpl, ping *pingTest) (*pongTest, Priority) {
		// t.Log("event:", *ping)
		return &pongTest{At: time.Now().UnixNano()}, Low
	})
	sess := NewSessionImpl(newConnTest(), &JSONProtocol{}, time.Second*30)
	b.Handle(sess)
}
