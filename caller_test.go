package bi

import "testing"
import "unsafe"

type SessionTest struct {
	Content string
}

type EventTest struct {
	Content string `json:"I,omitempty"`
}

type AckTest struct {
	Content string `json:"I,omitempty"`
}

func Benchmark_Caller(b *testing.B) {
	sess := &SessionTest{Content: "say "}
	sessPtr := unsafe.Pointer(sess)
	protocol := &JSONProtocol{}
	caller := newCaller(func(sess *SessionTest, event *EventTest) *AckTest {
		return &AckTest{Content: sess.Content + event.Content}
	})
	event := &EventTest{Content: "hello"}
	eventBytes, err := protocol.Marshal(event)
	if nil != err {
		b.Error(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := caller.call(sessPtr, protocol, eventBytes)
		if nil != err {
			b.Error(err)
		}
	}
	b.StopTimer()
}
