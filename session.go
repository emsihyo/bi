package bi

import (
	"sync"
	"time"
	"unsafe"
)

//Session Session
type Session interface {
	Conn() Conn
	Protocol() Protocol
	SendPayloadBytes(payloadBytes []byte, priority Priority) error
	Send(method string, argument interface{}, ack interface{}, priority Priority) error
	handle(bi BI)
}

const (
	//Normal Normal
	Normal Priority = iota
	//Low Low
	Low
	//High High
	High
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
	mut                  sync.Mutex
	willSendPayloadBytes [High + 1]chan []byte
	errorOccurred        chan error
	didDisconnects       []chan error
	timeout              time.Duration
}

//NewSessionImpl NewSessionImpl
func NewSessionImpl(conn Conn, protocol Protocol, timeout time.Duration) *SessionImpl {
	return &SessionImpl{hand: newHandler(), didDisconnects: []chan error{}, errorOccurred: make(chan error), willSendPayloadBytes: [High + 1]chan []byte{make(chan []byte, 128), make(chan []byte, 256), make(chan []byte, 512)}, conn: conn, protocol: protocol, timeout: timeout}
}

//Protocol Protocol
func (sess *SessionImpl) Protocol() Protocol {
	return sess.protocol
}

//Conn Conn
func (sess *SessionImpl) Conn() Conn {
	return sess.conn
}

//SendPayloadBytes SendPayloadBytes must confirm that the protocols match before calling
func (sess *SessionImpl) SendPayloadBytes(payloadBytes []byte, priority Priority) error {
	select {
	case sess.willSendPayloadBytes[priority] <- payloadBytes:
	default:
		select {
		case sess.errorOccurred <- ErrChanFull:
		default:
		}
		return ErrChanFull
	}
	return nil
}

/*
Send     Send
method   method name
argument argument,marshalled argument data or struct ptr
*/
func (sess *SessionImpl) Send(method string, argument interface{}, ack interface{}, priority Priority) error {
	var argumentBytes []byte
	var err error
	protocol := sess.protocol
	switch data := argument.(type) {
	case *[]byte:
		//marshalled bytes
		argumentBytes = *data
	default:
		//marshal payload argument
		if argumentBytes, err = protocol.Marshal(data); nil != err {
			return err
		}
	}
	payload := payloadPool.Get()
	defer payloadPool.Put(payload)
	payload.T = Type_Event
	payload.M = method
	payload.A = argumentBytes
	if nil == ack {
		//do not wait for ack
		return sess.sendPayload(payload, priority)
	}
	var waiting chan error
	//waiting for disconnection
	if waiting, err = sess.addWaiting(); nil != err {
		return err
	}
	defer sess.removeWaiting(waiting)
	hand := sess.hand
	callback := make(chan []byte, 1)
	payload.I = hand.addCall(callback)
	defer hand.removeCall(payload.I)
	timer := timerPool.Get(sess.timeout)
	defer timerPool.Put(timer)
	if err = sess.sendPayload(payload, priority); nil != err {
		return err
	}
	select {
	case err = <-waiting:
	case <-timer.C:
		err = ErrTimeOut
	case ackBytes := <-callback:
		switch data := ack.(type) {
		//bytes is working for transmiting
		case *[]byte:
			*data = append(*data, ackBytes...)
		default:
			err = protocol.Unmarshal(ackBytes, ack)
		}
	}
	return err
}

func (sess *SessionImpl) sendPayload(payload *Payload, priority Priority) error {
	payloadBytes, err := sess.protocol.Marshal(payload)
	if nil != err {
		return err
	}
	return sess.SendPayloadBytes(payloadBytes, priority)
}

func (sess *SessionImpl) addWaiting() (chan error, error) {
	sess.mut.Lock()
	defer sess.mut.Unlock()
	if nil != sess.err {
		return nil, sess.err
	}
	didDisconnect := make(chan error, 1)
	sess.didDisconnects = append(sess.didDisconnects, didDisconnect)
	return didDisconnect, nil
}

func (sess *SessionImpl) removeWaiting(waiting chan error) {
	sess.mut.Lock()
	defer sess.mut.Unlock()
	for i, w := range sess.didDisconnects {
		if w == waiting {
			sess.didDisconnects = append(sess.didDisconnects[:i], sess.didDisconnects[i+1:]...)
			return
		}
	}
}

func (sess *SessionImpl) handle(bi BI) {
	//send payload bytes loop
	go func() {
	SEND:
		for {
			var packageBytes []byte
			var ok bool
			select {
			case packageBytes, ok = <-sess.willSendPayloadBytes[High]:
				if !ok {
					break SEND
				}
				sess.conn.Write(packageBytes)
				continue
			default:
			}
			select {
			case packageBytes, ok = <-sess.willSendPayloadBytes[Normal]:
				if !ok {
					break SEND
				}
				sess.conn.Write(packageBytes)
				continue
			default:
			}
			select {
			case packageBytes, ok = <-sess.willSendPayloadBytes[Low]:
				if !ok {
					break SEND
				}
				sess.conn.Write(packageBytes)
			default:
			}
		}
	}()
	//receive payloadBytes loop
	payloadBytesReceived := make(chan []byte)
	go func() {
		var err error
		var payloadBytes []byte
	RECEIVE:
		for {
			if payloadBytes, err = sess.conn.Read(); nil != err {
				sess.errorOccurred <- err
				break RECEIVE
			}
			payloadBytesReceived <- payloadBytes
		}
	}()
	var err error
	var payloadBtyes []byte
	var ackBytes []byte
	priority := Normal
	timer := timerPool.Get(sess.timeout)
	defer timerPool.Put(timer)
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
			payload := payloadPool.Get()
			defer payloadPool.Put(payload)
			if err = sess.protocol.Unmarshal(payloadBtyes, payload); nil != err {
				break
			}
			switch payload.T {
			case Type_Event:
				ackBytes, priority, err = bi.OnEvent(sessPtr, payload.M, protocol, payload.A)
				if nil != err {
					break
				}
				if nil == ackBytes {
					break
				}
				payload.T = Type_Ack
				payload.A = ackBytes
				err = sess.sendPayload(payload, priority)
			case Type_Ack:
				callback := sess.hand.getCall(payload.I)
				if nil != callback {
					callback <- payload.A
				}
			}
		case err = <-sess.errorOccurred:
		}
		if nil != err {
			break
		}
	}
	sess.conn.Close()
FOR1:
	for {
		select {
		case <-payloadBytesReceived:
		default:
			break FOR1
		}
	}
FOR2:
	for {
		select {
		case <-sess.errorOccurred:
		default:
			break FOR2
		}
	}
	close(sess.willSendPayloadBytes[Normal])
	close(sess.willSendPayloadBytes[Low])
	close(sess.willSendPayloadBytes[High])
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
