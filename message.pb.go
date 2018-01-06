// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/emsihyo/bi/message.proto

/*
	Package bi is a generated protocol buffer package.

	It is generated from these files:
		github.com/emsihyo/bi/message.proto

	It has these top-level messages:
		Package
*/
package bi

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/gogo/protobuf/gogoproto"

import strings "strings"
import reflect "reflect"

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Type int32

const (
	Type_Event Type = 0
	Type_Ack   Type = 1
)

var Type_name = map[int32]string{
	0: "Event",
	1: "Ack",
}
var Type_value = map[string]int32{
	"Event": 0,
	"Ack":   1,
}

func (x Type) String() string {
	return proto.EnumName(Type_name, int32(x))
}
func (Type) EnumDescriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

type Package struct {
	I uint64 `protobuf:"varint,1,opt,name=I,proto3" json:"I,omitempty"`
	T Type   `protobuf:"varint,2,opt,name=T,proto3,enum=bi.Type" json:"T,omitempty"`
	M string `protobuf:"bytes,3,opt,name=M,proto3" json:"M,omitempty"`
	A []byte `protobuf:"bytes,4,opt,name=A,proto3" json:"A,omitempty"`
}

func (m *Package) Reset()                    { *m = Package{} }
func (*Package) ProtoMessage()               {}
func (*Package) Descriptor() ([]byte, []int) { return fileDescriptorMessage, []int{0} }

func (m *Package) GetI() uint64 {
	if m != nil {
		return m.I
	}
	return 0
}

func (m *Package) GetT() Type {
	if m != nil {
		return m.T
	}
	return Type_Event
}

func (m *Package) GetM() string {
	if m != nil {
		return m.M
	}
	return ""
}

func (m *Package) GetA() []byte {
	if m != nil {
		return m.A
	}
	return nil
}

func init() {
	proto.RegisterType((*Package)(nil), "bi.Package")
	proto.RegisterEnum("bi.Type", Type_name, Type_value)
}
func (m *Package) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Package) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.I != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.I))
	}
	if m.T != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintMessage(dAtA, i, uint64(m.T))
	}
	if len(m.M) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintMessage(dAtA, i, uint64(len(m.M)))
		i += copy(dAtA[i:], m.M)
	}
	if len(m.A) > 0 {
		dAtA[i] = 0x22
		i++
		i = encodeVarintMessage(dAtA, i, uint64(len(m.A)))
		i += copy(dAtA[i:], m.A)
	}
	return i, nil
}

func encodeFixed64Message(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Message(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintMessage(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Package) Size() (n int) {
	var l int
	_ = l
	if m.I != 0 {
		n += 1 + sovMessage(uint64(m.I))
	}
	if m.T != 0 {
		n += 1 + sovMessage(uint64(m.T))
	}
	l = len(m.M)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	l = len(m.A)
	if l > 0 {
		n += 1 + l + sovMessage(uint64(l))
	}
	return n
}

func sovMessage(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozMessage(x uint64) (n int) {
	return sovMessage(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *Package) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Package{`,
		`I:` + fmt.Sprintf("%v", this.I) + `,`,
		`T:` + fmt.Sprintf("%v", this.T) + `,`,
		`M:` + fmt.Sprintf("%v", this.M) + `,`,
		`A:` + fmt.Sprintf("%v", this.A) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringMessage(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *Package) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Package: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Package: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field I", wireType)
			}
			m.I = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.I |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field T", wireType)
			}
			m.T = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.T |= (Type(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field M", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.M = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field A", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMessage
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.A = append(m.A[:0], dAtA[iNdEx:postIndex]...)
			if m.A == nil {
				m.A = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMessage(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMessage
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMessage(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMessage
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMessage
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthMessage
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowMessage
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipMessage(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthMessage = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMessage   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("github.com/emsihyo/bi/message.proto", fileDescriptorMessage) }

var fileDescriptorMessage = []byte{
	// 237 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4e, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcd, 0x2d, 0xce, 0xcc, 0xa8, 0xcc, 0xd7, 0x4f,
	0xca, 0xd4, 0xcf, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c, 0x4f, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x4a, 0xca, 0x94, 0xd2, 0x45, 0x52, 0x98, 0x9e, 0x9f, 0x9e, 0xaf, 0x0f, 0x96, 0x4a, 0x2a,
	0x4d, 0x03, 0xf3, 0xc0, 0x1c, 0x30, 0x0b, 0xa2, 0x45, 0xc9, 0x9b, 0x8b, 0x3d, 0x20, 0x31, 0x39,
	0x3b, 0x31, 0x3d, 0x55, 0x88, 0x87, 0x8b, 0xd1, 0x53, 0x82, 0x51, 0x81, 0x51, 0x83, 0x25, 0x88,
	0xd1, 0x53, 0x48, 0x8c, 0x8b, 0x31, 0x44, 0x82, 0x49, 0x81, 0x51, 0x83, 0xcf, 0x88, 0x43, 0x2f,
	0x29, 0x53, 0x2f, 0xa4, 0xb2, 0x20, 0x35, 0x88, 0x31, 0x04, 0xa4, 0xca, 0x57, 0x82, 0x59, 0x81,
	0x51, 0x83, 0x33, 0x88, 0xd1, 0x17, 0xc4, 0x73, 0x94, 0x60, 0x51, 0x60, 0xd4, 0xe0, 0x09, 0x62,
	0x74, 0xd4, 0x92, 0xe2, 0x62, 0x01, 0x29, 0x13, 0xe2, 0xe4, 0x62, 0x75, 0x2d, 0x4b, 0xcd, 0x2b,
	0x11, 0x60, 0x10, 0x62, 0xe7, 0x62, 0x76, 0x4c, 0xce, 0x16, 0x60, 0x74, 0x32, 0xb8, 0xf1, 0x50,
	0x8e, 0xa1, 0xe1, 0x91, 0x1c, 0xe3, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78,
	0x24, 0xc7, 0xc8, 0xc5, 0x97, 0x9c, 0x9f, 0xab, 0x07, 0xf5, 0x8f, 0x5e, 0x52, 0xa6, 0x13, 0x93,
	0x93, 0xe7, 0x02, 0x46, 0xc6, 0x45, 0x4c, 0x4c, 0x4e, 0x9e, 0x49, 0x6c, 0x60, 0x17, 0x1a, 0x03,
	0x02, 0x00, 0x00, 0xff, 0xff, 0xda, 0xd2, 0xa4, 0xc5, 0xfb, 0x00, 0x00, 0x00,
}
