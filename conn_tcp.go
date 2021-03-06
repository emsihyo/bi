package bi

import (
	"encoding/binary"
	"net"
)

//TCPConn TCPConn
type TCPConn struct {
	conn            *net.TCPConn
	maximumBodySize uint32
}

//NewTCPConn NewTCPConn
func NewTCPConn(conn *net.TCPConn, maximumBodySize uint32) *TCPConn {
	return &TCPConn{conn: conn, maximumBodySize: maximumBodySize}
}

//Close Close
func (conn *TCPConn) Close() {
	conn.conn.Close()
}

//RemoteAddr RemoteAddr
func (conn *TCPConn) RemoteAddr() string {
	return conn.conn.RemoteAddr().String()
}

//Read Read
func (conn *TCPConn) Read() ([]byte, error) {
	head := make([]byte, 4)
	if err := conn.read(head); nil != err {
		return nil, err
	}
	bodySize := binary.BigEndian.Uint32(head)
	if 0 == bodySize {
		return nil, nil
	}
	if conn.maximumBodySize < bodySize {
		return nil, ErrTooLargePayload
	}
	data := make([]byte, bodySize)
	if err := conn.read(data); nil != err {
		return nil, err
	}
	return data, nil
}

//Write Write
func (conn *TCPConn) Write(data []byte) error {
	head := make([]byte, 4)
	if nil != data && 0 <= len(data) {
		binary.BigEndian.PutUint32(head, uint32(len(data)))
		head = append(head, data...)
	} else {
		binary.BigEndian.PutUint32(head, 0)
	}
	return conn.write(head)
}

func (conn *TCPConn) read(data []byte) error {
	didReadBytesTotal := 0
	didReadBytes := 0
	var err error
	for {
		if didReadBytesTotal == len(data) {
			break
		}
		if didReadBytes, err = conn.conn.Read(data[didReadBytesTotal:]); nil != err {
			// log.Println(err)
			break
		}
		didReadBytesTotal += didReadBytes
	}
	return err
}

func (conn *TCPConn) write(data []byte) error {
	var err error
	didWriteBytes := 0
	for 0 < len(data) {
		didWriteBytes, err = conn.conn.Write(data)
		if nil != err {
			break
		}
		data = data[didWriteBytes:]
	}
	return err
}
