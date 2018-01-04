package bi

import "testing"
import "unsafe"

type sessionTest struct {
}

type reqTest struct {
	Content string `json:"I,omitempty"`
}

type respTest struct {
	Content string `json:"I,omitempty"`
}

func Test_caller(t *testing.T) {
	sess := &sessionTest{}
	p := &JSONProtocol{}
	caller := newCaller(func(sess *sessionTest, in *reqTest) *respTest {
		return &respTest{Content: in.Content}
	})
	req := &reqTest{Content: "hello"}
	reqData, err := p.Marshal(req)
	if nil != err {
		t.Error(err)
	}
	respData, err := caller.call(unsafe.Pointer(sess), p, reqData)
	if nil != err {
		t.Error(err)
	}
	resp := &respTest{}
	err = p.Unmarshal(respData, resp)
	if nil != err {
		t.Error(err)
	}
}
