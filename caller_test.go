package bi

import "testing"
import "unsafe"

type SessionTest struct {
}

type EventTest struct {
	Content string `json:"I,omitempty"`
}

type AckTest struct {
	Content string `json:"I,omitempty"`
}

func Benchmark_Caller(b *testing.B) {
	sess := &SessionTest{}
	protocol := &JSONProtocol{}
	caller := newCaller(func(sess *SessionTest, event *EventTest) *AckTest {
		return &AckTest{Content: event.Content}
	})
	event := &EventTest{Content: "hello"}
	eventBytes, err := protocol.Marshal(event)
	if nil != err {
		b.Error(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := caller.call(unsafe.Pointer(sess), protocol, eventBytes)
		if nil != err {
			b.Error(err)
		}
	}
	b.StopTimer()
}
