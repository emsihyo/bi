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
	conn                   Conn
	protocol               Protocol
	hand                   *handler
	err                    error
	mut                    sync.Mutex
	didReceivePayloadBytes chan []byte
	willSendPayloadBytes   chan []byte
	didMakeError           chan error
	didDisconnects         []chan error
	timeout                time.Duration
	timer                  *time.Timer
}

//NewSession NewSession
func NewSession(conn Conn, protocol Protocol, timeout time.Duration) *Session {
	return &Session{hand: newHandler(), didDisconnects: []chan error{}, didReceivePayloadBytes: make(chan []byte), didMakeError: make(chan error, 1), willSendPayloadBytes: make(chan []byte, 256), conn: conn, protocol: protocol, timeout: timeout}
}

//GetProtocol GetProtocol
func (sess *Session) GetProtocol() Protocol {
	return sess.protocol
}

//Close Close
func (sess *Session) Close() {
	sess.conn.Close()
}

//SendPayloadBytes Should confirm that the protocols match.
func (sess *Session) SendPayloadBytes(payloadBytes []byte) {
	select {
	case sess.willSendPayloadBytes <- payloadBytes:
	default:
		select {
		case sess.didMakeError <- ErrChanFull:
		default:
		}
	}
}

//Send Send
func (sess *Session) Send(method string, argument interface{}, ack interface{}) error {
	var argumentBytes []byte
	var err error
	protocol := sess.protocol
	switch data := argument.(type) {
	case *[]byte:
		argumentBytes = *data
	default:
		if argumentBytes, err = protocol.Marshal(data); nil != err {
			return err
		}
	}
	payload := Pool.Payload.Get()
	defer Pool.Payload.Put(payload)
	payload.T = Type_Event
	payload.M = method
	payload.A = argumentBytes
	if nil == ack {
		return sess.sendPayload(payload)
	}
	var waiting chan error
	if waiting, err = sess.waiting(); nil != err {
		return err
	}
	hand := sess.hand
	payload.I = hand.nextCallID()
	callback := make(chan []byte, 1)
	hand.addCall(payload.I, callback)
	timer := Pool.Timer.Get(sess.timeout)
	defer hand.removeCall(payload.I)
	defer Pool.Timer.Put(timer)
	if err = sess.sendPayload(payload); nil != err {
		return err
	}
	select {
	case err = <-waiting:
	case <-timer.C:
		err = ErrTimeOut
	case ackBytes := <-callback:
		switch data := ack.(type) {
		case *[]byte:
			*data = append(*data, ackBytes...)
		default:
			err = protocol.Unmarshal(ackBytes, ack)
		}
	}
	return err
}

func (sess *Session) sendPayload(payload *Payload) error {
	payloadBytes, err := sess.protocol.Marshal(payload)
	if nil != err {
		return err
	}
	sess.SendPayloadBytes(payloadBytes)
	return nil
}

func (sess *Session) waiting() (chan error, error) {
	sess.mut.Lock()
	defer sess.mut.Unlock()
	if nil != sess.err {
		return nil, sess.err
	}
	didDisconnect := make(chan error, 1)
	sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	return didDisconnect, nil
}

func (sess *Session) sendLoop() {
	go func() {
		waiting, err := sess.waiting()
		if nil != err {
			return
		}
		for {
			select {
			case <-waiting:
				return
			case packageBytes := <-sess.willSendPayloadBytes:
				err = sess.conn.Write(packageBytes)
			}
			if nil != err {
				sess.conn.Close()
				return
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
				sess.didMakeError <- err
				break
			}
			sess.didReceivePayloadBytes <- data
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
		case payloadBtyes := <-sess.didReceivePayloadBytes:
			if nil != payloadBtyes {
				payload := Pool.Payload.Get()
				defer Pool.Payload.Put(payload)
				if err = sess.protocol.Unmarshal(payloadBtyes, payload); nil == err {
					switch payload.T {
					case Type_Event:
						ackBytes := []byte{}
						ackBytes, err = bi.onEvent(sessPtr, payload.M, protocol, payload.A)
						if nil == err {
							payload.T = Type_Ack
							payload.A = ackBytes
							err = sess.sendPayload(payload)
						}
					case Type_Ack:
						callback := sess.hand.getCall(payload.I)
						if nil != callback {
							callback <- payload.A
						}
					}
				}
			}
		case err = <-sess.didMakeError:
		}
		if nil != err {
			sess.conn.Close()
			break loop2
		}
	}
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
