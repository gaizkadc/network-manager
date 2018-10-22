// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tester.proto

package tester

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// MetadataType to test the enumerations.
type MetadataType int32

const (
	MetadataType_TYPE_A MetadataType = 0
	MetadataType_TYPE_B MetadataType = 1
)

var MetadataType_name = map[int32]string{
	0: "TYPE_A",
	1: "TYPE_B",
}
var MetadataType_value = map[string]int32{
	"TYPE_A": 0,
	"TYPE_B": 1,
}

func (x MetadataType) String() string {
	return proto.EnumName(MetadataType_name, int32(x))
}
func (MetadataType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_tester_94f34a97753751ca, []int{0}
}

// Metadata message with an enumeration and array of strings.
type Metadata struct {
	Type                 MetadataType `protobuf:"varint,1,opt,name=type,proto3,enum=tester.MetadataType" json:"type,omitempty"`
	Tags                 []string     `protobuf:"bytes,2,rep,name=tags,proto3" json:"tags,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Metadata) Reset()         { *m = Metadata{} }
func (m *Metadata) String() string { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()    {}
func (*Metadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_tester_94f34a97753751ca, []int{0}
}
func (m *Metadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Metadata.Unmarshal(m, b)
}
func (m *Metadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Metadata.Marshal(b, m, deterministic)
}
func (dst *Metadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Metadata.Merge(dst, src)
}
func (m *Metadata) XXX_Size() int {
	return xxx_messageInfo_Metadata.Size(m)
}
func (m *Metadata) XXX_DiscardUnknown() {
	xxx_messageInfo_Metadata.DiscardUnknown(m)
}

var xxx_messageInfo_Metadata proto.InternalMessageInfo

func (m *Metadata) GetType() MetadataType {
	if m != nil {
		return m.Type
	}
	return MetadataType_TYPE_A
}

func (m *Metadata) GetTags() []string {
	if m != nil {
		return m.Tags
	}
	return nil
}

// ComplexRequest linking to other messages.
type ComplexRequest struct {
	RequestNumber        int32     `protobuf:"varint,1,opt,name=request_number,json=requestNumber,proto3" json:"request_number,omitempty"`
	Name                 string    `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	InduceFailure        bool      `protobuf:"varint,3,opt,name=induce_failure,json=induceFailure,proto3" json:"induce_failure,omitempty"`
	Metadata             *Metadata `protobuf:"bytes,4,opt,name=metadata,proto3" json:"metadata,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ComplexRequest) Reset()         { *m = ComplexRequest{} }
func (m *ComplexRequest) String() string { return proto.CompactTextString(m) }
func (*ComplexRequest) ProtoMessage()    {}
func (*ComplexRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_tester_94f34a97753751ca, []int{1}
}
func (m *ComplexRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ComplexRequest.Unmarshal(m, b)
}
func (m *ComplexRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ComplexRequest.Marshal(b, m, deterministic)
}
func (dst *ComplexRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ComplexRequest.Merge(dst, src)
}
func (m *ComplexRequest) XXX_Size() int {
	return xxx_messageInfo_ComplexRequest.Size(m)
}
func (m *ComplexRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ComplexRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ComplexRequest proto.InternalMessageInfo

func (m *ComplexRequest) GetRequestNumber() int32 {
	if m != nil {
		return m.RequestNumber
	}
	return 0
}

func (m *ComplexRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ComplexRequest) GetInduceFailure() bool {
	if m != nil {
		return m.InduceFailure
	}
	return false
}

func (m *ComplexRequest) GetMetadata() *Metadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

// Response from the server
type Response struct {
	Msg                  string   `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	IsValid              bool     `protobuf:"varint,2,opt,name=is_valid,json=isValid,proto3" json:"is_valid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_tester_94f34a97753751ca, []int{2}
}
func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (dst *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(dst, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func (m *Response) GetIsValid() bool {
	if m != nil {
		return m.IsValid
	}
	return false
}

func init() {
	proto.RegisterType((*Metadata)(nil), "tester.Metadata")
	proto.RegisterType((*ComplexRequest)(nil), "tester.ComplexRequest")
	proto.RegisterType((*Response)(nil), "tester.Response")
	proto.RegisterEnum("tester.MetadataType", MetadataType_name, MetadataType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TesterClient is the client API for Tester service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TesterClient interface {
	// Process a complex request.
	ProcessComplexRequest(ctx context.Context, in *ComplexRequest, opts ...grpc.CallOption) (*Response, error)
}

type testerClient struct {
	cc *grpc.ClientConn
}

func NewTesterClient(cc *grpc.ClientConn) TesterClient {
	return &testerClient{cc}
}

func (c *testerClient) ProcessComplexRequest(ctx context.Context, in *ComplexRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/tester.Tester/ProcessComplexRequest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TesterServer is the server API for Tester service.
type TesterServer interface {
	// Process a complex request.
	ProcessComplexRequest(context.Context, *ComplexRequest) (*Response, error)
}

func RegisterTesterServer(s *grpc.Server, srv TesterServer) {
	s.RegisterService(&_Tester_serviceDesc, srv)
}

func _Tester_ProcessComplexRequest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ComplexRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TesterServer).ProcessComplexRequest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tester.Tester/ProcessComplexRequest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TesterServer).ProcessComplexRequest(ctx, req.(*ComplexRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Tester_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tester.Tester",
	HandlerType: (*TesterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessComplexRequest",
			Handler:    _Tester_ProcessComplexRequest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tester.proto",
}

func init() { proto.RegisterFile("tester.proto", fileDescriptor_tester_94f34a97753751ca) }

var fileDescriptor_tester_94f34a97753751ca = []byte{
	// 299 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x91, 0x4f, 0x4b, 0xc3, 0x40,
	0x10, 0xc5, 0xbb, 0x6d, 0x8d, 0xdb, 0xb1, 0x2d, 0x61, 0x51, 0x89, 0x9e, 0x42, 0x40, 0x09, 0x22,
	0x3d, 0xd4, 0x83, 0x67, 0x2d, 0x8a, 0x97, 0x4a, 0x59, 0x8a, 0xe0, 0x29, 0x6c, 0x9b, 0xb1, 0x04,
	0xf2, 0xcf, 0xdd, 0x8d, 0xd8, 0x6f, 0xe3, 0x47, 0x95, 0xdd, 0xfc, 0x41, 0x7b, 0x7b, 0xf3, 0x76,
	0xe6, 0xed, 0x8f, 0x19, 0x18, 0x6b, 0x54, 0x1a, 0xe5, 0xac, 0x94, 0x85, 0x2e, 0x98, 0x53, 0x57,
	0xc1, 0x0b, 0xd0, 0x25, 0x6a, 0x11, 0x0b, 0x2d, 0x58, 0x08, 0x43, 0xbd, 0x2f, 0xd1, 0x23, 0x3e,
	0x09, 0xa7, 0xf3, 0xd3, 0x59, 0x33, 0xd0, 0xbe, 0xaf, 0xf7, 0x25, 0x72, 0xdb, 0xc1, 0x18, 0x0c,
	0xb5, 0xd8, 0x29, 0xaf, 0xef, 0x0f, 0xc2, 0x11, 0xb7, 0x3a, 0xf8, 0x21, 0x30, 0x5d, 0x14, 0x59,
	0x99, 0xe2, 0x37, 0xc7, 0xcf, 0x0a, 0x95, 0x66, 0x57, 0x30, 0x95, 0xb5, 0x8c, 0xf2, 0x2a, 0xdb,
	0xa0, 0xb4, 0xd1, 0x47, 0x7c, 0xd2, 0xb8, 0xaf, 0xd6, 0x34, 0x69, 0xb9, 0xc8, 0xd0, 0xeb, 0xfb,
	0xc4, 0xa4, 0x19, 0x6d, 0x46, 0x93, 0x3c, 0xae, 0xb6, 0x18, 0x7d, 0x88, 0x24, 0xad, 0x24, 0x7a,
	0x03, 0x9f, 0x84, 0x94, 0x4f, 0x6a, 0xf7, 0xb9, 0x36, 0xd9, 0x2d, 0xd0, 0xac, 0xc1, 0xf3, 0x86,
	0x3e, 0x09, 0x4f, 0xe6, 0xee, 0x21, 0x36, 0xef, 0x3a, 0x82, 0x7b, 0xa0, 0x1c, 0x55, 0x59, 0xe4,
	0x0a, 0x99, 0x0b, 0x83, 0x4c, 0xed, 0x2c, 0xd0, 0x88, 0x1b, 0xc9, 0x2e, 0x80, 0x26, 0x2a, 0xfa,
	0x12, 0x69, 0x12, 0x5b, 0x14, 0xca, 0x8f, 0x13, 0xf5, 0x66, 0xca, 0x9b, 0x6b, 0x18, 0xff, 0xdd,
	0x02, 0x03, 0x70, 0xd6, 0xef, 0xab, 0xa7, 0xe8, 0xc1, 0xed, 0x75, 0xfa, 0xd1, 0x25, 0xf3, 0x25,
	0x38, 0x6b, 0xfb, 0x3b, 0x5b, 0xc0, 0xd9, 0x4a, 0x16, 0x5b, 0x54, 0xea, 0x60, 0x27, 0xe7, 0x2d,
	0xdf, 0x7f, 0xff, 0xb2, 0xe3, 0x6e, 0x09, 0x83, 0xde, 0xc6, 0xb1, 0xb7, 0xba, 0xfb, 0x0d, 0x00,
	0x00, 0xff, 0xff, 0x61, 0x21, 0x4e, 0x71, 0xbb, 0x01, 0x00, 0x00,
}
