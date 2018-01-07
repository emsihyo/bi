package bi

import (
	"errors"
	"sync"
	"time"
	"unsafe"
)

var (
	//ErrChanFull ErrChanFull
	ErrChanFull = errors.New("bi.session chan is full")
	//ErrTimeOut ErrTimeOut
	ErrTimeOut = errors.New("bi.session timeout")
	//ErrClosed ErrClosed
	ErrClosed = errors.New("bi.session closed")
)

//Session Session
const (
	//Connection Connection
	Connection = "_CONNECTION"
	//Disconnection Disconnection
	Disconnection = "_DISCONNECTION"
)

//Session Session
type Session struct {
	id                string
	conn              Conn
	protocol          Protocol
	hand              *handler
	err               error
	mut               sync.Mutex
	marshalledEvent   chan []byte
	didReceivePackage chan *Package
	didReceiveError   chan error
	didDisconnects    []chan error
	timeout           time.Duration
	timer             *time.Timer
}

//NewSession NewSession
func NewSession(conn Conn, protocol Protocol, timeout time.Duration) *Session {
	return &Session{hand: newHandler(), didDisconnects: []chan error{}, didReceivePackage: make(chan *Package), didReceiveError: make(chan error, 1), marshalledEvent: make(chan []byte, 512), conn: conn, protocol: protocol, timeout: timeout}
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

//MarshalledEvent MarshalledEvent
func (sess *Session) MarshalledEvent(marshalledEvent []byte) {
	select {
	case sess.marshalledEvent <- marshalledEvent:
	default:
		select {
		case sess.didReceiveError <- ErrChanFull:
		default:
		}
	}
}

//Event Event
func (sess *Session) Event(method string, event interface{}, ack interface{}) (err error) {
	var eventBytes []byte
	protocol := sess.protocol
	if eventBytes, err = protocol.Marshal(event); nil != err {
		// log.Println(err)
		return err
	}
	pkg := Pool.pkg.Get()
	defer Pool.pkg.Put(pkg)
	pkg.T = Type_Event
	pkg.M = method
	pkg.A = eventBytes
	if nil == ack {
		sess.sendEvent(pkg)
		return err
	}
	waiting := sess.waiting()
	if nil != waiting {
		hand := sess.hand
		pkg.I = hand.nextCallID()
		callback := make(chan []byte, 1)
		hand.addCall(pkg.I, callback)
		timer := Pool.Timer.Get(sess.timeout)
		sess.sendEvent(pkg)
		defer func() {
			hand.removeCall(pkg.I)
			Pool.Timer.Put(timer)
		}()
		select {
		case <-timer.C:
			return ErrTimeOut
		case ackBytes := <-callback:
			if err = protocol.Unmarshal(ackBytes, ack); nil != err {
				return err
			}
			return nil
		case err = <-waiting:
			return err
		}
	} else {
		return sess.err
	}
}

func (sess *Session) sendEvent(pkg *Package) {
	marshalledEvent, err := sess.protocol.Marshal(pkg)
	if nil != err {
		// log.Println(err)
		return
	}
	sess.marshalledEvent <- marshalledEvent
}

func (sess *Session) waiting() chan error {
	sess.mut.Lock()
	defer sess.mut.Unlock()
	if nil != sess.err {
		return nil
	}
	didDisconnect := make(chan error, 1)
	sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	return didDisconnect
}

func (sess *Session) sendLoop() {
	go func() {
		waiting := sess.waiting()
		if nil != waiting {
			for {
				select {
				case <-waiting:
					return
				case packageBytes := <-sess.marshalledEvent:
					sess.conn.Write(packageBytes)
				}
			}
		}
	}()
}

func (sess *Session) receiveLoop() {
	go func() {
		var err error
		var data []byte
		for {
			if data, err = sess.conn.Read(); nil != err {
				sess.didReceiveError <- err
				break
			}
			if nil == data {
				sess.didReceivePackage <- nil
				continue
			}
			pkg := Pool.pkg.Get()
			if err = sess.protocol.Unmarshal(data, pkg); nil != err {
				sess.didReceiveError <- err
				Pool.pkg.Put(pkg)
				break
			} else {
				sess.didReceivePackage <- pkg
				Pool.pkg.Put(pkg)
			}
		}
	}()
}
func (sess *Session) handle(bi *BI) {
	sess.sendLoop()
	sess.receiveLoop()
	var err error
	sessPtr := unsafe.Pointer(sess)
	protocol := sess.protocol
	bi.onEvent(sessPtr, Connection, protocol, nil)
loop2:
	for {
		sess.timer.Reset(sess.timeout)
		select {
		case <-sess.timer.C:
			err = ErrTimeOut
			sess.conn.Close()
		case pkg := <-sess.didReceivePackage:
			if nil != err && nil != pkg {
				switch pkg.T {
				case Type_Event:
					ackBytes := []byte{}
					ackBytes, err = bi.onEvent(sessPtr, pkg.M, protocol, pkg.A)
					if nil == err {
						pkg.T = Type_Ack
						pkg.A = ackBytes
						sess.sendEvent(pkg)
					}
				case Type_Ack:
					if nil != err {
						callback := sess.hand.getCall(pkg.I)
						if nil != callback {
							callback <- pkg.A
						}
					}
				}
			}
		case err = <-sess.didReceiveError:
			break loop2
		}
	}
	sess.conn.Close()
	sess.mut.Lock()
	defer sess.mut.Unlock()
	if nil == sess.err {
		sess.err = err
	}
	for _, didDisconnect := range sess.didDisconnects {
		didDisconnect <- err
	}
	bi.onEvent(sessPtr, Disconnection, protocol, nil)
}
