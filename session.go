package bi

import (
	"errors"
	"sync"
	"time"
	"unsafe"
)

//Session Session
type Session interface {
	Close()
	handle(bi BI)
}

var (
	//ErrTooLargePayload ErrTooLargePayload
	ErrTooLargePayload = errors.New("bi.session pay load is too large")
	//ErrChanFull ErrChanFull
	ErrChanFull = errors.New("bi.session chan is full")
	//ErrTimeOut ErrTimeOut
	ErrTimeOut = errors.New("bi.session timeout")
	//ErrClosed ErrClosed
	ErrClosed = errors.New("bi.session closed")
)

//Weight Weight
type Weight int

const (

	//Normal Normal
	Normal Weight = iota
	//Lazy Lazy
	Lazy
	//Urgent Urgent
	Urgent
)

//SessionImpl SessionImpl
const (
	//Connection Connection
	Connection = "_CONNECTION"
	//Disconnection Disconnection
	Disconnection = "_DISCONNECTION"
)

//SessionImpl SessionImpl
type SessionImpl struct {
	conn                 Conn
	protocol             Protocol
	hand                 *handler
	err                  error
	mut                  *sync.Mutex
	willSendPayloadBytes [Urgent + 1]chan []byte
	errorOccurred        chan error
	didDisconnects       []chan error
	timeout              time.Duration
}

//NewSessionImpl NewSessionImpl
func NewSessionImpl(conn Conn, protocol Protocol, timeout time.Duration) *SessionImpl {
	return &SessionImpl{mut: &sync.Mutex{}, hand: newHandler(), didDisconnects: []chan error{}, errorOccurred: make(chan error, 1), willSendPayloadBytes: [Urgent + 1]chan []byte{make(chan []byte, 128), make(chan []byte, 256), make(chan []byte, 512)}, conn: conn, protocol: protocol, timeout: timeout}
}

//GetProtocol GetProtocol
func (sess *SessionImpl) GetProtocol() Protocol {
	return sess.protocol
}

//Close Close
func (sess *SessionImpl) Close() {
	sess.conn.Close()
}

//SendPayloadBytes Should confirm that the protocols match.
func (sess *SessionImpl) SendPayloadBytes(payloadBytes []byte, weight Weight) error {
	select {
	case sess.willSendPayloadBytes[weight] <- payloadBytes:
	default:
		sess.errorOccurred <- ErrChanFull
		return ErrChanFull
	}
	return nil
}

//Send Send
func (sess *SessionImpl) Send(method string, argument interface{}, ack interface{}, weight Weight) error {
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
		return sess.sendPayload(payload, weight)
	}
	var waiting chan error
	if waiting, err = sess.waiting(); nil != err {
		return err
	}
	hand := sess.hand
	callback := make(chan []byte, 1)
	payload.I = hand.addCall(callback)
	timer := Pool.Timer.Get(sess.timeout)
	defer hand.removeCall(payload.I)
	defer Pool.Timer.Put(timer)
	if err = sess.sendPayload(payload, weight); nil != err {
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

func (sess *SessionImpl) sendPayload(payload *Payload, weight Weight) error {
	payloadBytes, err := sess.protocol.Marshal(payload)
	if nil != err {
		return err
	}
	return sess.SendPayloadBytes(payloadBytes, weight)
}

func (sess *SessionImpl) waiting() (chan error, error) {
	sess.mut.Lock()
	defer sess.mut.Unlock()
	if nil != sess.err {
		return nil, sess.err
	}
	didDisconnect := make(chan error, 1)
	sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	return didDisconnect, nil
}

//Handle Handle
func (sess *SessionImpl) handle(bi BI) {
	payloadBytesReceived := make(chan []byte)
	//send payload bytes loop
	go func() {
		var err error
		for {
			var packageBytes []byte
			var ok bool
			select {
			case packageBytes, ok = <-sess.willSendPayloadBytes[Urgent]:
				if !ok {
					return
				}
				err = sess.conn.Write(packageBytes)
				if nil != err {
					sess.errorOccurred <- err
					return
				}
				continue
			default:
			}
			select {
			case packageBytes, ok = <-sess.willSendPayloadBytes[Normal]:
				if !ok {
					return
				}
				err = sess.conn.Write(packageBytes)
				if nil != err {
					sess.errorOccurred <- err
					return
				}
				continue
			default:
			}
			select {
			case packageBytes, ok = <-sess.willSendPayloadBytes[Lazy]:
				if !ok {
					return
				}
				err = sess.conn.Write(packageBytes)
				if nil != err {
					sess.errorOccurred <- err
					return
				}
			default:
			}
		}
	}()
	//receive payloadBytes loop
	go func() {
		var err error
		var payloadBytes []byte
		for {
			if payloadBytes, err = sess.conn.Read(); nil != err {
				sess.errorOccurred <- err
				continue
			}
			payloadBytesReceived <- payloadBytes
		}
	}()
	var err error
	var payloadBtyes []byte
	var ackBytes []byte
	timer := Pool.Timer.Get(sess.timeout)
	defer Pool.Timer.Put(timer)
	sessPtr := unsafe.Pointer(sess)
	protocol := sess.protocol
	bi.OnEvent(sessPtr, Connection, protocol, nil)
	for {
		timer.Reset(sess.timeout)
		select {
		case <-timer.C:
			err = ErrTimeOut
		case payloadBtyes = <-payloadBytesReceived:
			if nil == payloadBtyes {
				break
			}
			payload := Pool.Payload.Get()
			defer Pool.Payload.Put(payload)
			if err = sess.protocol.Unmarshal(payloadBtyes, payload); nil != err {
				break
			}
			switch payload.T {
			case Type_Event:
				ackBytes, err = bi.OnEvent(sessPtr, payload.M, protocol, payload.A)
				if nil != err {
					break
				}
				if nil == ackBytes {
					break
				}
				payload.T = Type_Ack
				payload.A = ackBytes
				err = sess.sendPayload(payload, Normal)
			case Type_Ack:
				callback := sess.hand.getCall(payload.I)
				if nil != callback {
					callback <- payload.A
				}
			}
		case err = <-sess.errorOccurred:
		}
		if nil != err {
			sess.conn.Close()
			close(payloadBytesReceived)
			close(sess.willSendPayloadBytes[Normal])
			close(sess.willSendPayloadBytes[Lazy])
			close(sess.willSendPayloadBytes[Urgent])
			close(sess.errorOccurred)
			break
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
	bi.OnEvent(sessPtr, Disconnection, protocol, nil)
}
