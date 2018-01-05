package bi

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	uuid "github.com/satori/go.uuid"
)

var (
	//ErrTimeOut ErrTimeOut
	ErrTimeOut = errors.New("bi.session timeout")
	//ErrClosed ErrClosed
	ErrClosed = errors.New("bi.session closed")
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
	id                    string
	conn                  Conn
	protocol              Protocol
	hand                  *handler
	isClosed              bool
	closedErr             error
	closedMut             sync.Mutex
	sendMarshalledPackage chan []byte
	didReceivePackage     chan *Package
	didReceiveError       chan error
	didDisconnects        []chan error
	timeout               time.Duration
	timer                 *time.Timer
}

//NewSession NewSession
func NewSession(conn Conn, protocol Protocol, timeout time.Duration) *Session {
	return &Session{id: uuid.NewV3(uuid.NewV4(), conn.RemoteAddr()).String(), conn: conn, protocol: protocol, hand: newHandler(), didDisconnects: []chan error{}, didReceivePackage: make(chan *Package, 64), didReceiveError: make(chan error, 1), sendMarshalledPackage: make(chan []byte, 512), timeout: timeout, timer: time.NewTimer(timeout)}
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

//SendMarshalledPackage SendMarshalledPackage
func (sess *Session) SendMarshalledPackage(marshalledData []byte) {
	sess.sendMarshalledPackage <- marshalledData
}

//Event Event
func (sess *Session) Event(method string, event interface{}, ack interface{}) (err error) {
	var eventBytes []byte
	protocol := sess.protocol
	if eventBytes, err = protocol.Marshal(event); nil != err {
		// log.Println(err)
		return err
	}
	pkg := &Package{}
	pkg.T = Type_Event
	pkg.M = method
	pkg.A = eventBytes
	if nil == ack {
		sess.sendPackage(pkg)
		return err
	}
	waitForDisconnection := sess.waitForDisconnection()
	if nil != waitForDisconnection {
		hand := sess.hand
		id := hand.nextCallID()
		pkg.I = id
		callback := make(chan []byte, 1)
		hand.addCall(id, callback)
		defer hand.removeCall(pkg.I)
		sess.sendPackage(pkg)

		select {
		case <-time.After(sess.timeout):
			return ErrTimeOut
		case ackBytes := <-callback:
			if err = protocol.Unmarshal(ackBytes, ack); nil != err {
				return err
			}
			return nil
		case err = <-waitForDisconnection:
			return err
		}
	} else {
		return sess.closedErr
	}
}

func (sess *Session) sendPackage(pkg *Package) {
	marshalledData, err := sess.protocol.Marshal(pkg)
	if nil != err {
		// log.Println(err)
		return
	}
	sess.sendMarshalledPackage <- marshalledData
}

func (sess *Session) waitForDisconnection() chan error {
	sess.closedMut.Lock()
	defer sess.closedMut.Unlock()
	if true == sess.isClosed {
		return nil
	}
	didDisconnect := make(chan error, 1)
	sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	return didDisconnect
}

func (sess *Session) sendLoop() {
	go func() {
		waitForDisconnection := sess.waitForDisconnection()
		if nil != waitForDisconnection {
			for {
				select {
				case <-waitForDisconnection:
					return
				case packageBytes := <-sess.sendMarshalledPackage:
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
			pkg := Package{}
			if err = sess.protocol.Unmarshal(data, &pkg); nil != err {
				sess.didReceiveError <- err
				break
			} else {
				sess.didReceivePackage <- &pkg
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
						sess.sendPackage(pkg)
					} else {
						pkg.T = Type_Error
						pkg.A = []byte(err.Error())
						sess.sendPackage(pkg)
						sess.conn.Close()
					}
				case Type_Ack:
					if nil != err {
						callback := sess.hand.getCall(pkg.I)
						if nil != callback {
							callback <- pkg.A
						}
					}
				case Type_Error:
					if nil != err {
						err = errors.New(string(pkg.A))
						sess.conn.Close()
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
	bi.onEvent(sessPtr, Disconnection, protocol, nil)
}
