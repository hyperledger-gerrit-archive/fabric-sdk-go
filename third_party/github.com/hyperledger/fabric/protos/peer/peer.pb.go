// Code generated by protoc-gen-go. DO NOT EDIT.
// source: peer/peer.proto

package peer

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

type PeerID struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *PeerID) Reset()                    { *m = PeerID{} }
func (m *PeerID) String() string            { return proto.CompactTextString(m) }
func (*PeerID) ProtoMessage()               {}
func (*PeerID) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{0} }

func (m *PeerID) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type PeerEndpoint struct {
	Id      *PeerID `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Address string  `protobuf:"bytes,2,opt,name=address" json:"address,omitempty"`
}

func (m *PeerEndpoint) Reset()                    { *m = PeerEndpoint{} }
func (m *PeerEndpoint) String() string            { return proto.CompactTextString(m) }
func (*PeerEndpoint) ProtoMessage()               {}
func (*PeerEndpoint) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{1} }

func (m *PeerEndpoint) GetId() *PeerID {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *PeerEndpoint) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func init() {
	proto.RegisterType((*PeerID)(nil), "protos.PeerID")
	proto.RegisterType((*PeerEndpoint)(nil), "protos.PeerEndpoint")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Endorser service

type EndorserClient interface {
	ProcessProposal(ctx context.Context, in *SignedProposal, opts ...grpc.CallOption) (*ProposalResponse, error)
}

type endorserClient struct {
	cc *grpc.ClientConn
}

func NewEndorserClient(cc *grpc.ClientConn) EndorserClient {
	return &endorserClient{cc}
}

func (c *endorserClient) ProcessProposal(ctx context.Context, in *SignedProposal, opts ...grpc.CallOption) (*ProposalResponse, error) {
	out := new(ProposalResponse)
	err := grpc.Invoke(ctx, "/protos.Endorser/ProcessProposal", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Endorser service

type EndorserServer interface {
	ProcessProposal(context.Context, *SignedProposal) (*ProposalResponse, error)
}

func RegisterEndorserServer(s *grpc.Server, srv EndorserServer) {
	s.RegisterService(&_Endorser_serviceDesc, srv)
}

func _Endorser_ProcessProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignedProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EndorserServer).ProcessProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Endorser/ProcessProposal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EndorserServer).ProcessProposal(ctx, req.(*SignedProposal))
	}
	return interceptor(ctx, in, info, handler)
}

var _Endorser_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Endorser",
	HandlerType: (*EndorserServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessProposal",
			Handler:    _Endorser_ProcessProposal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "peer/peer.proto",
}

// Client API for Channel service

type ChannelClient interface {
	Chat(ctx context.Context, opts ...grpc.CallOption) (Channel_ChatClient, error)
}

type channelClient struct {
	cc *grpc.ClientConn
}

func NewChannelClient(cc *grpc.ClientConn) ChannelClient {
	return &channelClient{cc}
}

func (c *channelClient) Chat(ctx context.Context, opts ...grpc.CallOption) (Channel_ChatClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Channel_serviceDesc.Streams[0], c.cc, "/protos.Channel/Chat", opts...)
	if err != nil {
		return nil, err
	}
	x := &channelChatClient{stream}
	return x, nil
}

type Channel_ChatClient interface {
	Send(*SignedEvent) error
	Recv() (*Event, error)
	grpc.ClientStream
}

type channelChatClient struct {
	grpc.ClientStream
}

func (x *channelChatClient) Send(m *SignedEvent) error {
	return x.ClientStream.SendMsg(m)
}

func (x *channelChatClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Channel service

type ChannelServer interface {
	Chat(Channel_ChatServer) error
}

func RegisterChannelServer(s *grpc.Server, srv ChannelServer) {
	s.RegisterService(&_Channel_serviceDesc, srv)
}

func _Channel_Chat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChannelServer).Chat(&channelChatServer{stream})
}

type Channel_ChatServer interface {
	Send(*Event) error
	Recv() (*SignedEvent, error)
	grpc.ServerStream
}

type channelChatServer struct {
	grpc.ServerStream
}

func (x *channelChatServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func (x *channelChatServer) Recv() (*SignedEvent, error) {
	m := new(SignedEvent)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Channel_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Channel",
	HandlerType: (*ChannelServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Chat",
			Handler:       _Channel_Chat_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "peer/peer.proto",
}

func init() { proto.RegisterFile("peer/peer.proto", fileDescriptor6) }

var fileDescriptor6 = []byte{
	// 289 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x91, 0xcf, 0x4b, 0xfb, 0x40,
	0x10, 0xc5, 0x9b, 0x52, 0xda, 0xef, 0x77, 0xfd, 0x51, 0xdc, 0x82, 0x84, 0x50, 0x44, 0x72, 0xaa,
	0x97, 0xa4, 0xd4, 0xa3, 0x37, 0x6b, 0x40, 0x4f, 0xc6, 0x78, 0xf3, 0x22, 0x9b, 0xec, 0x98, 0x2c,
	0xa4, 0xbb, 0xcb, 0x4c, 0x14, 0xfc, 0xef, 0x25, 0xbb, 0x89, 0xd8, 0x4b, 0xb2, 0xf3, 0xde, 0x67,
	0xdf, 0x0e, 0x33, 0x6c, 0x69, 0x01, 0x30, 0xed, 0x3f, 0x89, 0x45, 0xd3, 0x19, 0x3e, 0x77, 0x3f,
	0x8a, 0x2e, 0x9c, 0x01, 0x5f, 0xa0, 0x3b, 0xf2, 0x56, 0xb4, 0xf2, 0x2c, 0x1a, 0x6b, 0x48, 0xb4,
	0x83, 0xb8, 0x3e, 0x12, 0xdf, 0x11, 0xc8, 0x1a, 0x4d, 0xe0, 0xdd, 0x78, 0xcd, 0xe6, 0x39, 0x00,
	0x3e, 0x3d, 0x70, 0xce, 0x66, 0x5a, 0x1c, 0x20, 0x0c, 0xae, 0x83, 0xcd, 0xff, 0xc2, 0x9d, 0xe3,
	0x47, 0x76, 0xda, 0xbb, 0x99, 0x96, 0xd6, 0x28, 0xdd, 0xf1, 0x2b, 0x36, 0x55, 0xd2, 0x11, 0x27,
	0xbb, 0x73, 0x9f, 0x40, 0x89, 0xbf, 0x5f, 0x4c, 0x95, 0xe4, 0x21, 0x5b, 0x08, 0x29, 0x11, 0x88,
	0xc2, 0xa9, 0x8b, 0x19, 0xcb, 0xdd, 0x0b, 0xfb, 0x97, 0x69, 0x69, 0x90, 0x00, 0x79, 0xc6, 0x96,
	0x39, 0x9a, 0x0a, 0x88, 0xf2, 0xa1, 0x2b, 0x7e, 0x39, 0x86, 0xbd, 0xaa, 0x5a, 0x83, 0x1c, 0xf5,
	0x28, 0xfc, 0x7d, 0x64, 0x50, 0x8a, 0xa1, 0xfd, 0x78, 0xb2, 0xbb, 0x63, 0x8b, 0x7d, 0x23, 0xb4,
	0x86, 0x96, 0x6f, 0xd9, 0x6c, 0xdf, 0x88, 0x8e, 0xaf, 0x8e, 0x63, 0xb2, 0x7e, 0x38, 0xd1, 0xd9,
	0x28, 0xba, 0x32, 0x9e, 0x6c, 0x82, 0x6d, 0x70, 0xff, 0xcc, 0x62, 0x83, 0x75, 0xd2, 0x7c, 0x5b,
	0xc0, 0x16, 0x64, 0x0d, 0x98, 0x7c, 0x88, 0x12, 0x55, 0x35, 0xc2, 0xfd, 0xd4, 0xde, 0x6e, 0x6a,
	0xd5, 0x35, 0x9f, 0x65, 0x52, 0x99, 0x43, 0xfa, 0x07, 0x4d, 0x3d, 0x9a, 0x7a, 0xd4, 0x2d, 0xa7,
	0xf4, 0x6b, 0xb9, 0xfd, 0x09, 0x00, 0x00, 0xff, 0xff, 0xe3, 0x62, 0xb0, 0x68, 0xb0, 0x01, 0x00,
	0x00,
}
