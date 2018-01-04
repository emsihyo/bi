package bi

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	uuid "github.com/satori/go.uuid"
)

//Session Session
const (
	//Connection Connection
	Connection = "Connection"
	//Disconnection Disconnection
	Disconnection = "Disconnection"
)

//Session Session
type Session struct {
	id                 string
	protocol           Protocol
	conn               Conn
	qa                 *QA
	isClosed           bool
	closedErr          error
	closedMut          sync.Mutex
	sendMarshalledData chan []byte
	didReceiveMessage  chan *Message
	didReceiveError    chan error
	didDisconnects     []chan error
	timeout            time.Duration
	timer              *time.Timer
}

//NewSession NewSession
func NewSession(conn Conn, protocol Protocol, timeout time.Duration) *Session {
	return &Session{id: uuid.NewV3(uuid.NewV4(), conn.RemoteAddr()).String(), conn: conn, protocol: protocol, qa: newQA(), didDisconnects: []chan error{}, didReceiveMessage: make(chan *Message, 64), didReceiveError: make(chan error, 1), sendMarshalledData: make(chan []byte, 512), timeout: timeout, timer: time.NewTimer(timeout)}
}

//GetID GetID
func (sess *Session) GetID() string {
	return sess.id
}

//GetProtocol GetProtocol
func (sess *Session) GetProtocol() Protocol {
	return sess.protocol
}

//Close Close
func (sess *Session) Close() {
	sess.conn.Close()
}

//Emit Emit
func (sess *Session) Emit(method string, req interface{}) error {
	var a []byte
	var err error
	if nil != req {
		if a, err = sess.protocol.Marshal(req); nil != err {
			// log.Println(err)
			return err
		}
	}
	m := Message{
		T: Message_Emit,
		M: method,
		A: a,
	}
	sess.sendMessage(&m)
	return nil
}

//Request Request
func (sess *Session) Request(method string, req interface{}, resp interface{}) error {
	var a []byte
	var err error
	protocol := sess.protocol
	if nil != req {
		if a, err = protocol.Marshal(req); nil != err {
			// log.Println(err)
			return err
		}
	}
	qa := sess.qa
	m := Message{
		T: Message_Request,
		I: qa.nextID(),
		M: method,
		A: a,
	}
	q := make(chan []byte, 1)
	qa.addQ(m.I, q)
	defer qa.removeQ(m.I)
	waitForDisconnection := sess.waitForDisconnection()
	sess.sendMessage(&m)
	select {
	case respBytes := <-q:
		// err = protocol.Unmarshal(respBytes, resp)
		// if nil != err {
		// 	log.Println(err)
		// }
		return protocol.Unmarshal(respBytes, resp)
	case err = <-waitForDisconnection:
		if nil != err {
			return err
		}
		return errors.New("bi: session.impl.closed")
	}
}

//SendMarshalledData SendMarshalledData
func (sess *Session) SendMarshalledData(marshalledData []byte) {
	sess.sendMarshalledData <- marshalledData
}

func (sess *Session) handle(bi *BI) {
	sessPtr := unsafe.Pointer(sess)
	bi.onRequest(sessPtr, Connection, sess.protocol, nil)
	waitForDisconnection := sess.waitForDisconnection()
	//receive
	go func() {
		var err error
		var data []byte
		for {
			if data, err = sess.conn.Read(); nil != err {
				sess.didReceiveError <- err
				break
			}
			if nil == data {
				sess.didReceiveMessage <- nil
				continue
			}
			m := Message{}
			err = sess.protocol.Unmarshal(data, &m)
			if nil != err {
				// log.Println(err)
				sess.didReceiveError <- err
				break
			} else {
				sess.didReceiveMessage <- &m
			}
		}
	}()

	go func() {
		//send loop
	loop1:
		for {
			select {
			case <-waitForDisconnection:
				break loop1
			case data := <-sess.sendMarshalledData:
				sess.conn.Write(data)
			}
		}
	}()

	//receive loop
	var err error
	var didTimeout bool
loop2:
	for {
		sess.timer.Reset(sess.timeout)
		select {
		case <-sess.timer.C:
			//timeout
			didTimeout = true
			sess.conn.Close()
		case m := <-sess.didReceiveMessage:
			if true != didTimeout && nil != m {
				switch m.T {
				case Message_Emit:
					bi.onRequest(sessPtr, m.M, sess.protocol, m.A)
				case Message_Request:
					resp, _ := bi.onRequest(sessPtr, m.M, sess.protocol, m.A)
					m.T = Message_Response
					m.A = resp
					m.M = ""
					sess.sendMessage(m)
				case Message_Response:
					q := sess.qa.getQ(m.I)
					if nil != q {
						q <- m.A
					}
				}
			}
		case err = <-sess.didReceiveError:
			break loop2
		}
	}
	sess.closedMut.Lock()
	defer sess.closedMut.Unlock()
	sess.closedErr = err
	sess.isClosed = true
	sess.conn.Close()
	for _, didDisconnect := range sess.didDisconnects {
		didDisconnect <- err
	}
	bi.onRequest(sessPtr, Disconnection, sess.protocol, nil)
}

func (sess *Session) sendMessage(m *Message) {
	marshalledData, err := sess.protocol.Marshal(m)
	if nil != err {
		// log.Println(err)
		return
	}
	sess.sendMarshalledData <- marshalledData
}

func (sess *Session) waitForDisconnection() chan error {
	didDisconnect := make(chan error, 1)
	sess.closedMut.Lock()
	defer sess.closedMut.Unlock()
	if true == sess.isClosed {
		go func() {
			didDisconnect <- sess.closedErr
		}()
	} else {
		sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	}
	return didDisconnect
}
