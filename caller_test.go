package bi

import "testing"
import "unsafe"

type sessionTest struct {
	Content string
}

type eventTest struct {
	Content string `json:"I,omitempty"`
}

type ackTest struct {
	Content string `json:"I,omitempty"`
}

func Benchmark_Caller(b *testing.B) {
	sess := &sessionTest{Content: "say "}
	sessPtr := unsafe.Pointer(sess)
	protocol := &JSONProtocol{}
	caller := newCaller(func(sess *sessionTest, event *eventTest) (*ackTest, Weight) {
		return &ackTest{Content: sess.Content + event.Content}, Urgent
	})
	event := &eventTest{Content: "hello"}
	eventBytes, err := protocol.Marshal(event)
	if nil != err {
		b.Error(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := caller.call(sessPtr, protocol, eventBytes)
		if nil != err {
			b.Error(err)
		}
	}
	b.StopTimer()
}
