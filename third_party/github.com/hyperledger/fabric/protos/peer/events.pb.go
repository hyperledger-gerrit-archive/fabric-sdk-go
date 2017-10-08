/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/
// Code generated by protoc-gen-go. DO NOT EDIT.
// source: peer/events.proto

package peer

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import common "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type EventType int32

const (
	EventType_REGISTER        EventType = 0
	EventType_BLOCK           EventType = 1
	EventType_CHAINCODE       EventType = 2
	EventType_REJECTION       EventType = 3
	EventType_FILTEREDBLOCK   EventType = 4
	EventType_REGISTERCHANNEL EventType = 5
)

var EventType_name = map[int32]string{
	0: "REGISTER",
	1: "BLOCK",
	2: "CHAINCODE",
	3: "REJECTION",
	4: "FILTEREDBLOCK",
	5: "REGISTERCHANNEL",
}
var EventType_value = map[string]int32{
	"REGISTER":        0,
	"BLOCK":           1,
	"CHAINCODE":       2,
	"REJECTION":       3,
	"FILTEREDBLOCK":   4,
	"REGISTERCHANNEL": 5,
}

func (x EventType) String() string {
	return proto.EnumName(EventType_name, int32(x))
}
func (EventType) EnumDescriptor() ([]byte, []int) { return fileDescriptor5, []int{0} }

// ChaincodeReg is used for registering chaincode Interests
// when EventType is CHAINCODE
type ChaincodeReg struct {
	ChaincodeId string `protobuf:"bytes,1,opt,name=chaincode_id,json=chaincodeId" json:"chaincode_id,omitempty"`
	EventName   string `protobuf:"bytes,2,opt,name=event_name,json=eventName" json:"event_name,omitempty"`
}

func (m *ChaincodeReg) Reset()                    { *m = ChaincodeReg{} }
func (m *ChaincodeReg) String() string            { return proto.CompactTextString(m) }
func (*ChaincodeReg) ProtoMessage()               {}
func (*ChaincodeReg) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{0} }

func (m *ChaincodeReg) GetChaincodeId() string {
	if m != nil {
		return m.ChaincodeId
	}
	return ""
}

func (m *ChaincodeReg) GetEventName() string {
	if m != nil {
		return m.EventName
	}
	return ""
}

type Interest struct {
	EventType EventType `protobuf:"varint,1,opt,name=event_type,json=eventType,enum=protos.EventType" json:"event_type,omitempty"`
	// Ideally we should just have the following oneof for different
	// Reg types and get rid of EventType. But this is an API change
	// Additional Reg types may add messages specific to their type
	// to the oneof.
	//
	// Types that are valid to be assigned to RegInfo:
	//	*Interest_ChaincodeRegInfo
	RegInfo isInterest_RegInfo `protobuf_oneof:"RegInfo"`
	ChainID string             `protobuf:"bytes,3,opt,name=chainID" json:"chainID,omitempty"`
}

func (m *Interest) Reset()                    { *m = Interest{} }
func (m *Interest) String() string            { return proto.CompactTextString(m) }
func (*Interest) ProtoMessage()               {}
func (*Interest) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{1} }

type isInterest_RegInfo interface {
	isInterest_RegInfo()
}

type Interest_ChaincodeRegInfo struct {
	ChaincodeRegInfo *ChaincodeReg `protobuf:"bytes,2,opt,name=chaincode_reg_info,json=chaincodeRegInfo,oneof"`
}

func (*Interest_ChaincodeRegInfo) isInterest_RegInfo() {}

func (m *Interest) GetRegInfo() isInterest_RegInfo {
	if m != nil {
		return m.RegInfo
	}
	return nil
}

func (m *Interest) GetEventType() EventType {
	if m != nil {
		return m.EventType
	}
	return EventType_REGISTER
}

func (m *Interest) GetChaincodeRegInfo() *ChaincodeReg {
	if x, ok := m.GetRegInfo().(*Interest_ChaincodeRegInfo); ok {
		return x.ChaincodeRegInfo
	}
	return nil
}

func (m *Interest) GetChainID() string {
	if m != nil {
		return m.ChainID
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Interest) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Interest_OneofMarshaler, _Interest_OneofUnmarshaler, _Interest_OneofSizer, []interface{}{
		(*Interest_ChaincodeRegInfo)(nil),
	}
}

func _Interest_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Interest)
	// RegInfo
	switch x := m.RegInfo.(type) {
	case *Interest_ChaincodeRegInfo:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.ChaincodeRegInfo); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Interest.RegInfo has unexpected type %T", x)
	}
	return nil
}

func _Interest_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Interest)
	switch tag {
	case 2: // RegInfo.chaincode_reg_info
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(ChaincodeReg)
		err := b.DecodeMessage(msg)
		m.RegInfo = &Interest_ChaincodeRegInfo{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Interest_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Interest)
	// RegInfo
	switch x := m.RegInfo.(type) {
	case *Interest_ChaincodeRegInfo:
		s := proto.Size(x.ChaincodeRegInfo)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// ---------- consumer events ---------
// Register is sent by consumers for registering events
// string type - "register"
type Register struct {
	Events []*Interest `protobuf:"bytes,1,rep,name=events" json:"events,omitempty"`
}

func (m *Register) Reset()                    { *m = Register{} }
func (m *Register) String() string            { return proto.CompactTextString(m) }
func (*Register) ProtoMessage()               {}
func (*Register) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{2} }

func (m *Register) GetEvents() []*Interest {
	if m != nil {
		return m.Events
	}
	return nil
}

// Rejection is sent by consumers for erroneous transaction rejection events
// string type - "rejection"
type Rejection struct {
	Tx       *Transaction `protobuf:"bytes,1,opt,name=tx" json:"tx,omitempty"`
	ErrorMsg string       `protobuf:"bytes,2,opt,name=error_msg,json=errorMsg" json:"error_msg,omitempty"`
}

func (m *Rejection) Reset()                    { *m = Rejection{} }
func (m *Rejection) String() string            { return proto.CompactTextString(m) }
func (*Rejection) ProtoMessage()               {}
func (*Rejection) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{3} }

func (m *Rejection) GetTx() *Transaction {
	if m != nil {
		return m.Tx
	}
	return nil
}

func (m *Rejection) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

// ---------- producer events ---------
type Unregister struct {
	Events []*Interest `protobuf:"bytes,1,rep,name=events" json:"events,omitempty"`
}

func (m *Unregister) Reset()                    { *m = Unregister{} }
func (m *Unregister) String() string            { return proto.CompactTextString(m) }
func (*Unregister) ProtoMessage()               {}
func (*Unregister) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{4} }

func (m *Unregister) GetEvents() []*Interest {
	if m != nil {
		return m.Events
	}
	return nil
}

// FilteredBlock is sent by producers and contains minimal information
// about the block.
type FilteredBlock struct {
	ChannelId  string                 `protobuf:"bytes,1,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
	Number     uint64                 `protobuf:"varint,2,opt,name=number" json:"number,omitempty"`
	FilteredTx []*FilteredTransaction `protobuf:"bytes,3,rep,name=filtered_tx,json=filteredTx" json:"filtered_tx,omitempty"`
}

func (m *FilteredBlock) Reset()                    { *m = FilteredBlock{} }
func (m *FilteredBlock) String() string            { return proto.CompactTextString(m) }
func (*FilteredBlock) ProtoMessage()               {}
func (*FilteredBlock) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{5} }

func (m *FilteredBlock) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func (m *FilteredBlock) GetNumber() uint64 {
	if m != nil {
		return m.Number
	}
	return 0
}

func (m *FilteredBlock) GetFilteredTx() []*FilteredTransaction {
	if m != nil {
		return m.FilteredTx
	}
	return nil
}

// FilteredTransaction is a minimal set of information about a transaction
// within a block.
type FilteredTransaction struct {
	Txid             string           `protobuf:"bytes,1,opt,name=txid" json:"txid,omitempty"`
	TxValidationCode TxValidationCode `protobuf:"varint,2,opt,name=tx_validation_code,json=txValidationCode,enum=protos.TxValidationCode" json:"tx_validation_code,omitempty"`
	CcEvent          *ChaincodeEvent  `protobuf:"bytes,3,opt,name=ccEvent" json:"ccEvent,omitempty"`
}

func (m *FilteredTransaction) Reset()                    { *m = FilteredTransaction{} }
func (m *FilteredTransaction) String() string            { return proto.CompactTextString(m) }
func (*FilteredTransaction) ProtoMessage()               {}
func (*FilteredTransaction) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{6} }

func (m *FilteredTransaction) GetTxid() string {
	if m != nil {
		return m.Txid
	}
	return ""
}

func (m *FilteredTransaction) GetTxValidationCode() TxValidationCode {
	if m != nil {
		return m.TxValidationCode
	}
	return TxValidationCode_VALID
}

func (m *FilteredTransaction) GetCcEvent() *ChaincodeEvent {
	if m != nil {
		return m.CcEvent
	}
	return nil
}

// SignedEvent is used for any communication between consumer and producer
type SignedEvent struct {
	// Signature over the event bytes
	Signature []byte `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	// Marshal of Event object
	EventBytes []byte `protobuf:"bytes,2,opt,name=eventBytes,proto3" json:"eventBytes,omitempty"`
}

func (m *SignedEvent) Reset()                    { *m = SignedEvent{} }
func (m *SignedEvent) String() string            { return proto.CompactTextString(m) }
func (*SignedEvent) ProtoMessage()               {}
func (*SignedEvent) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{7} }

func (m *SignedEvent) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *SignedEvent) GetEventBytes() []byte {
	if m != nil {
		return m.EventBytes
	}
	return nil
}

// Event is used by
//  - consumers (adapters) to send Register
//  - producer to advertise supported types and events
type Event struct {
	// Types that are valid to be assigned to Event:
	//	*Event_Register
	//	*Event_Block
	//	*Event_ChaincodeEvent
	//	*Event_Rejection
	//	*Event_Unregister
	//	*Event_FilteredBlock
	//	*Event_RegisterChannel
	//	*Event_DeregisterChannel
	//	*Event_ChannelServiceResponse
	Event isEvent_Event `protobuf_oneof:"Event"`
	// Creator of the event, specified as a certificate chain
	Creator []byte `protobuf:"bytes,6,opt,name=creator,proto3" json:"creator,omitempty"`
	// Channel the event pertains to - used by the channel service when sending
	// block and filtered block events
	ChannelId string `protobuf:"bytes,11,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (m *Event) String() string            { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{8} }

type isEvent_Event interface {
	isEvent_Event()
}

type Event_Register struct {
	Register *Register `protobuf:"bytes,1,opt,name=register,oneof"`
}
type Event_Block struct {
	Block *common.Block `protobuf:"bytes,2,opt,name=block,oneof"`
}
type Event_ChaincodeEvent struct {
	ChaincodeEvent *ChaincodeEvent `protobuf:"bytes,3,opt,name=chaincode_event,json=chaincodeEvent,oneof"`
}
type Event_Rejection struct {
	Rejection *Rejection `protobuf:"bytes,4,opt,name=rejection,oneof"`
}
type Event_Unregister struct {
	Unregister *Unregister `protobuf:"bytes,5,opt,name=unregister,oneof"`
}
type Event_FilteredBlock struct {
	FilteredBlock *FilteredBlock `protobuf:"bytes,7,opt,name=filtered_block,json=filteredBlock,oneof"`
}
type Event_RegisterChannel struct {
	RegisterChannel *RegisterChannel `protobuf:"bytes,8,opt,name=register_channel,json=registerChannel,oneof"`
}
type Event_DeregisterChannel struct {
	DeregisterChannel *DeregisterChannel `protobuf:"bytes,9,opt,name=deregister_channel,json=deregisterChannel,oneof"`
}
type Event_ChannelServiceResponse struct {
	ChannelServiceResponse *ChannelServiceResponse `protobuf:"bytes,10,opt,name=channel_service_response,json=channelServiceResponse,oneof"`
}

func (*Event_Register) isEvent_Event()               {}
func (*Event_Block) isEvent_Event()                  {}
func (*Event_ChaincodeEvent) isEvent_Event()         {}
func (*Event_Rejection) isEvent_Event()              {}
func (*Event_Unregister) isEvent_Event()             {}
func (*Event_FilteredBlock) isEvent_Event()          {}
func (*Event_RegisterChannel) isEvent_Event()        {}
func (*Event_DeregisterChannel) isEvent_Event()      {}
func (*Event_ChannelServiceResponse) isEvent_Event() {}

func (m *Event) GetEvent() isEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (m *Event) GetRegister() *Register {
	if x, ok := m.GetEvent().(*Event_Register); ok {
		return x.Register
	}
	return nil
}

func (m *Event) GetBlock() *common.Block {
	if x, ok := m.GetEvent().(*Event_Block); ok {
		return x.Block
	}
	return nil
}

func (m *Event) GetChaincodeEvent() *ChaincodeEvent {
	if x, ok := m.GetEvent().(*Event_ChaincodeEvent); ok {
		return x.ChaincodeEvent
	}
	return nil
}

func (m *Event) GetRejection() *Rejection {
	if x, ok := m.GetEvent().(*Event_Rejection); ok {
		return x.Rejection
	}
	return nil
}

func (m *Event) GetUnregister() *Unregister {
	if x, ok := m.GetEvent().(*Event_Unregister); ok {
		return x.Unregister
	}
	return nil
}

func (m *Event) GetFilteredBlock() *FilteredBlock {
	if x, ok := m.GetEvent().(*Event_FilteredBlock); ok {
		return x.FilteredBlock
	}
	return nil
}

func (m *Event) GetRegisterChannel() *RegisterChannel {
	if x, ok := m.GetEvent().(*Event_RegisterChannel); ok {
		return x.RegisterChannel
	}
	return nil
}

func (m *Event) GetDeregisterChannel() *DeregisterChannel {
	if x, ok := m.GetEvent().(*Event_DeregisterChannel); ok {
		return x.DeregisterChannel
	}
	return nil
}

func (m *Event) GetChannelServiceResponse() *ChannelServiceResponse {
	if x, ok := m.GetEvent().(*Event_ChannelServiceResponse); ok {
		return x.ChannelServiceResponse
	}
	return nil
}

func (m *Event) GetCreator() []byte {
	if m != nil {
		return m.Creator
	}
	return nil
}

func (m *Event) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Event) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Event_OneofMarshaler, _Event_OneofUnmarshaler, _Event_OneofSizer, []interface{}{
		(*Event_Register)(nil),
		(*Event_Block)(nil),
		(*Event_ChaincodeEvent)(nil),
		(*Event_Rejection)(nil),
		(*Event_Unregister)(nil),
		(*Event_FilteredBlock)(nil),
		(*Event_RegisterChannel)(nil),
		(*Event_DeregisterChannel)(nil),
		(*Event_ChannelServiceResponse)(nil),
	}
}

func _Event_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Event)
	// Event
	switch x := m.Event.(type) {
	case *Event_Register:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Register); err != nil {
			return err
		}
	case *Event_Block:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Block); err != nil {
			return err
		}
	case *Event_ChaincodeEvent:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.ChaincodeEvent); err != nil {
			return err
		}
	case *Event_Rejection:
		b.EncodeVarint(4<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Rejection); err != nil {
			return err
		}
	case *Event_Unregister:
		b.EncodeVarint(5<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Unregister); err != nil {
			return err
		}
	case *Event_FilteredBlock:
		b.EncodeVarint(7<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.FilteredBlock); err != nil {
			return err
		}
	case *Event_RegisterChannel:
		b.EncodeVarint(8<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.RegisterChannel); err != nil {
			return err
		}
	case *Event_DeregisterChannel:
		b.EncodeVarint(9<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DeregisterChannel); err != nil {
			return err
		}
	case *Event_ChannelServiceResponse:
		b.EncodeVarint(10<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.ChannelServiceResponse); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Event.Event has unexpected type %T", x)
	}
	return nil
}

func _Event_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Event)
	switch tag {
	case 1: // Event.register
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Register)
		err := b.DecodeMessage(msg)
		m.Event = &Event_Register{msg}
		return true, err
	case 2: // Event.block
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(common.Block)
		err := b.DecodeMessage(msg)
		m.Event = &Event_Block{msg}
		return true, err
	case 3: // Event.chaincode_event
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(ChaincodeEvent)
		err := b.DecodeMessage(msg)
		m.Event = &Event_ChaincodeEvent{msg}
		return true, err
	case 4: // Event.rejection
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Rejection)
		err := b.DecodeMessage(msg)
		m.Event = &Event_Rejection{msg}
		return true, err
	case 5: // Event.unregister
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Unregister)
		err := b.DecodeMessage(msg)
		m.Event = &Event_Unregister{msg}
		return true, err
	case 7: // Event.filtered_block
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(FilteredBlock)
		err := b.DecodeMessage(msg)
		m.Event = &Event_FilteredBlock{msg}
		return true, err
	case 8: // Event.register_channel
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RegisterChannel)
		err := b.DecodeMessage(msg)
		m.Event = &Event_RegisterChannel{msg}
		return true, err
	case 9: // Event.deregister_channel
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(DeregisterChannel)
		err := b.DecodeMessage(msg)
		m.Event = &Event_DeregisterChannel{msg}
		return true, err
	case 10: // Event.channel_service_response
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(ChannelServiceResponse)
		err := b.DecodeMessage(msg)
		m.Event = &Event_ChannelServiceResponse{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Event_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Event)
	// Event
	switch x := m.Event.(type) {
	case *Event_Register:
		s := proto.Size(x.Register)
		n += proto.SizeVarint(1<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_Block:
		s := proto.Size(x.Block)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_ChaincodeEvent:
		s := proto.Size(x.ChaincodeEvent)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_Rejection:
		s := proto.Size(x.Rejection)
		n += proto.SizeVarint(4<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_Unregister:
		s := proto.Size(x.Unregister)
		n += proto.SizeVarint(5<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_FilteredBlock:
		s := proto.Size(x.FilteredBlock)
		n += proto.SizeVarint(7<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_RegisterChannel:
		s := proto.Size(x.RegisterChannel)
		n += proto.SizeVarint(8<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_DeregisterChannel:
		s := proto.Size(x.DeregisterChannel)
		n += proto.SizeVarint(9<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Event_ChannelServiceResponse:
		s := proto.Size(x.ChannelServiceResponse)
		n += proto.SizeVarint(10<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type RegisterChannel struct {
	ChannelIds []string    `protobuf:"bytes,1,rep,name=channel_ids,json=channelIds" json:"channel_ids,omitempty"`
	Events     []*Interest `protobuf:"bytes,2,rep,name=events" json:"events,omitempty"`
}

func (m *RegisterChannel) Reset()                    { *m = RegisterChannel{} }
func (m *RegisterChannel) String() string            { return proto.CompactTextString(m) }
func (*RegisterChannel) ProtoMessage()               {}
func (*RegisterChannel) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{9} }

func (m *RegisterChannel) GetChannelIds() []string {
	if m != nil {
		return m.ChannelIds
	}
	return nil
}

func (m *RegisterChannel) GetEvents() []*Interest {
	if m != nil {
		return m.Events
	}
	return nil
}

type DeregisterChannel struct {
	ChannelIds []string `protobuf:"bytes,1,rep,name=channel_ids,json=channelIds" json:"channel_ids,omitempty"`
}

func (m *DeregisterChannel) Reset()                    { *m = DeregisterChannel{} }
func (m *DeregisterChannel) String() string            { return proto.CompactTextString(m) }
func (*DeregisterChannel) ProtoMessage()               {}
func (*DeregisterChannel) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{10} }

func (m *DeregisterChannel) GetChannelIds() []string {
	if m != nil {
		return m.ChannelIds
	}
	return nil
}

// ChannelServiceResponse returns information about registration/deregistration
// actions on the server to the client. The possible actions are currently
// RegisterChannel and DeregisterChannel. Success indicates whether the action
// succeeded for all channels.
type ChannelServiceResponse struct {
	Action                string                  `protobuf:"bytes,1,opt,name=action" json:"action,omitempty"`
	Success               bool                    `protobuf:"varint,2,opt,name=success" json:"success,omitempty"`
	ChannelServiceResults []*ChannelServiceResult `protobuf:"bytes,3,rep,name=channel_service_results,json=channelServiceResults" json:"channel_service_results,omitempty"`
}

func (m *ChannelServiceResponse) Reset()                    { *m = ChannelServiceResponse{} }
func (m *ChannelServiceResponse) String() string            { return proto.CompactTextString(m) }
func (*ChannelServiceResponse) ProtoMessage()               {}
func (*ChannelServiceResponse) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{11} }

func (m *ChannelServiceResponse) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *ChannelServiceResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *ChannelServiceResponse) GetChannelServiceResults() []*ChannelServiceResult {
	if m != nil {
		return m.ChannelServiceResults
	}
	return nil
}

// ChannelServiceResult holds information about each action that was requested by
// the client. authorized_events holds the events the client has access to based
// on any ACL that is present. An empty error message means that the action was
// successful. Otherwise, it will contain context about the reason for failure.
type ChannelServiceResult struct {
	ChannelId        string   `protobuf:"bytes,1,opt,name=channel_id,json=channelId" json:"channel_id,omitempty"`
	AuthorizedEvents []string `protobuf:"bytes,2,rep,name=authorized_events,json=authorizedEvents" json:"authorized_events,omitempty"`
	ErrorMsg         string   `protobuf:"bytes,3,opt,name=error_msg,json=errorMsg" json:"error_msg,omitempty"`
}

func (m *ChannelServiceResult) Reset()                    { *m = ChannelServiceResult{} }
func (m *ChannelServiceResult) String() string            { return proto.CompactTextString(m) }
func (*ChannelServiceResult) ProtoMessage()               {}
func (*ChannelServiceResult) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{12} }

func (m *ChannelServiceResult) GetChannelId() string {
	if m != nil {
		return m.ChannelId
	}
	return ""
}

func (m *ChannelServiceResult) GetAuthorizedEvents() []string {
	if m != nil {
		return m.AuthorizedEvents
	}
	return nil
}

func (m *ChannelServiceResult) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

func init() {
	proto.RegisterType((*ChaincodeReg)(nil), "protos.ChaincodeReg")
	proto.RegisterType((*Interest)(nil), "protos.Interest")
	proto.RegisterType((*Register)(nil), "protos.Register")
	proto.RegisterType((*Rejection)(nil), "protos.Rejection")
	proto.RegisterType((*Unregister)(nil), "protos.Unregister")
	proto.RegisterType((*FilteredBlock)(nil), "protos.FilteredBlock")
	proto.RegisterType((*FilteredTransaction)(nil), "protos.FilteredTransaction")
	proto.RegisterType((*SignedEvent)(nil), "protos.SignedEvent")
	proto.RegisterType((*Event)(nil), "protos.Event")
	proto.RegisterType((*RegisterChannel)(nil), "protos.RegisterChannel")
	proto.RegisterType((*DeregisterChannel)(nil), "protos.DeregisterChannel")
	proto.RegisterType((*ChannelServiceResponse)(nil), "protos.ChannelServiceResponse")
	proto.RegisterType((*ChannelServiceResult)(nil), "protos.ChannelServiceResult")
	proto.RegisterEnum("protos.EventType", EventType_name, EventType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Events service

type EventsClient interface {
	// event chatting using Event
	Chat(ctx context.Context, opts ...grpc.CallOption) (Events_ChatClient, error)
}

type eventsClient struct {
	cc *grpc.ClientConn
}

func NewEventsClient(cc *grpc.ClientConn) EventsClient {
	return &eventsClient{cc}
}

func (c *eventsClient) Chat(ctx context.Context, opts ...grpc.CallOption) (Events_ChatClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Events_serviceDesc.Streams[0], c.cc, "/protos.Events/Chat", opts...)
	if err != nil {
		return nil, err
	}
	x := &eventsChatClient{stream}
	return x, nil
}

type Events_ChatClient interface {
	Send(*SignedEvent) error
	Recv() (*Event, error)
	grpc.ClientStream
}

type eventsChatClient struct {
	grpc.ClientStream
}

func (x *eventsChatClient) Send(m *SignedEvent) error {
	return x.ClientStream.SendMsg(m)
}

func (x *eventsChatClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Events service

type EventsServer interface {
	// event chatting using Event
	Chat(Events_ChatServer) error
}

func RegisterEventsServer(s *grpc.Server, srv EventsServer) {
	s.RegisterService(&_Events_serviceDesc, srv)
}

func _Events_Chat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EventsServer).Chat(&eventsChatServer{stream})
}

type Events_ChatServer interface {
	Send(*Event) error
	Recv() (*SignedEvent, error)
	grpc.ServerStream
}

type eventsChatServer struct {
	grpc.ServerStream
}

func (x *eventsChatServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func (x *eventsChatServer) Recv() (*SignedEvent, error) {
	m := new(SignedEvent)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Events_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Events",
	HandlerType: (*EventsServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Chat",
			Handler:       _Events_Chat_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "peer/events.proto",
}

func init() { proto.RegisterFile("peer/events.proto", fileDescriptor5) }

var fileDescriptor5 = []byte{
	// 993 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x56, 0xdf, 0x6f, 0xe3, 0x44,
	0x10, 0xb6, 0xdb, 0x34, 0x89, 0x27, 0x49, 0xeb, 0x6c, 0xef, 0x72, 0xa6, 0x77, 0x1c, 0x87, 0x11,
	0x52, 0x01, 0x29, 0x29, 0xa1, 0xe2, 0x01, 0x21, 0xa4, 0xe6, 0x47, 0x71, 0xee, 0x7a, 0xe9, 0x69,
	0x1b, 0x78, 0x38, 0x21, 0x22, 0xc7, 0xde, 0x38, 0xbe, 0x4b, 0xec, 0x68, 0xbd, 0xa9, 0x52, 0x5e,
	0x78, 0xe1, 0x0f, 0x41, 0xe2, 0x81, 0xff, 0x91, 0x27, 0xe4, 0xf5, 0xae, 0xed, 0xa4, 0x81, 0xea,
	0x9e, 0xec, 0x9d, 0xf9, 0xbe, 0xd9, 0xd9, 0x99, 0x6f, 0xc7, 0x86, 0xfa, 0x92, 0x10, 0xda, 0x22,
	0xb7, 0x24, 0x60, 0x51, 0x73, 0x49, 0x43, 0x16, 0xa2, 0x22, 0x7f, 0x44, 0x27, 0xc7, 0x4e, 0xb8,
	0x58, 0x84, 0x41, 0x2b, 0x79, 0x24, 0xce, 0x93, 0x13, 0x8e, 0x77, 0x66, 0xb6, 0x1f, 0x38, 0xa1,
	0x4b, 0xc6, 0x9c, 0x29, 0x7c, 0x0d, 0xee, 0x63, 0xd4, 0x0e, 0x22, 0xdb, 0x61, 0xbe, 0xe4, 0x98,
	0x6f, 0xa0, 0xda, 0x95, 0x04, 0x4c, 0x3c, 0xf4, 0x29, 0x54, 0xb3, 0x00, 0xbe, 0x6b, 0xa8, 0x2f,
	0xd4, 0x53, 0x0d, 0x57, 0x52, 0xdb, 0xc0, 0x45, 0x1f, 0x03, 0xf0, 0xc8, 0xe3, 0xc0, 0x5e, 0x10,
	0x63, 0x8f, 0x03, 0x34, 0x6e, 0x19, 0xda, 0x0b, 0x62, 0xfe, 0xad, 0x42, 0x79, 0x10, 0x30, 0x42,
	0x49, 0xc4, 0xd0, 0x99, 0xc4, 0xb2, 0xbb, 0x25, 0xe1, 0xc1, 0x0e, 0xdb, 0xf5, 0x64, 0xeb, 0xa8,
	0xd9, 0x8f, 0x3d, 0xa3, 0xbb, 0x25, 0x11, 0xf4, 0xf8, 0x15, 0xf5, 0x00, 0x65, 0x09, 0x50, 0xe2,
	0x8d, 0xfd, 0x60, 0x1a, 0xf2, 0x5d, 0x2a, 0xed, 0x47, 0x92, 0x99, 0x4f, 0xd9, 0x52, 0xb0, 0xee,
	0xe4, 0xd6, 0x83, 0x60, 0x1a, 0x22, 0x03, 0x4a, 0xdc, 0x36, 0xe8, 0x19, 0xfb, 0x3c, 0x41, 0xb9,
	0xec, 0x68, 0x50, 0x12, 0x20, 0xf3, 0x1c, 0xca, 0x98, 0x78, 0x7e, 0xc4, 0x08, 0x45, 0xa7, 0x50,
	0x4c, 0x0a, 0x6d, 0xa8, 0x2f, 0xf6, 0x4f, 0x2b, 0x6d, 0x5d, 0x6e, 0x25, 0x8f, 0x82, 0x85, 0xdf,
	0x7c, 0x0d, 0x1a, 0x26, 0xef, 0x08, 0x2f, 0x22, 0xfa, 0x0c, 0xf6, 0xd8, 0x9a, 0x9f, 0xab, 0xd2,
	0x3e, 0x96, 0x94, 0x51, 0x56, 0x65, 0xbc, 0xc7, 0xd6, 0xe8, 0x29, 0x68, 0x84, 0xd2, 0x90, 0x8e,
	0x17, 0x91, 0x27, 0xea, 0x55, 0xe6, 0x86, 0xd7, 0x91, 0x67, 0x7e, 0x0b, 0xf0, 0x53, 0x40, 0x3f,
	0x3c, 0x8d, 0x3f, 0x54, 0xa8, 0x5d, 0xfa, 0xf3, 0xd8, 0xea, 0x76, 0xe6, 0xa1, 0xf3, 0x3e, 0xee,
	0x8b, 0x33, 0xb3, 0x83, 0x80, 0xcc, 0xb3, 0xc6, 0x69, 0xc2, 0x32, 0x70, 0x51, 0x03, 0x8a, 0xc1,
	0x6a, 0x31, 0x21, 0x94, 0xa7, 0x50, 0xc0, 0x62, 0x85, 0xbe, 0x87, 0xca, 0x54, 0xc4, 0x19, 0xb3,
	0xb5, 0xb1, 0xcf, 0xf7, 0x7d, 0x2a, 0xf7, 0x95, 0x5b, 0xe4, 0xcf, 0x04, 0x12, 0x3f, 0x5a, 0x9b,
	0x7f, 0xa9, 0x70, 0xbc, 0x03, 0x83, 0x10, 0x14, 0xd8, 0x3a, 0x4d, 0x83, 0xbf, 0xa3, 0x4b, 0x40,
	0x6c, 0x3d, 0xbe, 0xb5, 0xe7, 0xbe, 0x6b, 0xc7, 0xa0, 0x71, 0xdc, 0x31, 0x9e, 0xcd, 0x61, 0xdb,
	0x48, 0x8b, 0xb7, 0xfe, 0x39, 0x05, 0x74, 0xe3, 0x8e, 0xea, 0x6c, 0xcb, 0x82, 0xce, 0xa0, 0xe4,
	0x38, 0x5c, 0x3c, 0xbc, 0xb9, 0x95, 0x76, 0xe3, 0x9e, 0x2e, 0xb8, 0x17, 0x4b, 0x98, 0xf9, 0x0a,
	0x2a, 0x37, 0xbe, 0x17, 0x10, 0x97, 0x2f, 0xd1, 0x33, 0xd0, 0x22, 0xdf, 0x0b, 0x6c, 0xb6, 0xa2,
	0x89, 0x28, 0xab, 0x38, 0x33, 0xa0, 0xe7, 0x42, 0xb3, 0x9d, 0x3b, 0x46, 0x22, 0x9e, 0x5e, 0x15,
	0xe7, 0x2c, 0xe6, 0x3f, 0x05, 0x38, 0x48, 0xe2, 0x34, 0xa1, 0x2c, 0x3b, 0x27, 0x34, 0x90, 0xf6,
	0x4b, 0x0a, 0xcb, 0x52, 0x70, 0x8a, 0x41, 0x9f, 0xc3, 0xc1, 0x24, 0x6e, 0x95, 0x90, 0x73, 0xad,
	0x29, 0xae, 0x2f, 0xef, 0x9f, 0xa5, 0xe0, 0xc4, 0x8b, 0x2e, 0xe0, 0x68, 0xeb, 0x12, 0xff, 0xff,
	0x39, 0x2d, 0x05, 0x1f, 0x3a, 0x1b, 0x16, 0xf4, 0x35, 0x68, 0x54, 0x8a, 0xd4, 0x28, 0x70, 0x72,
	0x3d, 0x4b, 0x4d, 0x38, 0x2c, 0x05, 0x67, 0x28, 0x74, 0x0e, 0xb0, 0x4a, 0x85, 0x68, 0x1c, 0x70,
	0x0e, 0x92, 0x9c, 0x4c, 0xa2, 0x96, 0x82, 0x73, 0x38, 0xf4, 0x03, 0x1c, 0xa6, 0xea, 0x49, 0xce,
	0x56, 0xe2, 0xcc, 0xc7, 0xdb, 0x02, 0x92, 0x67, 0xac, 0x4d, 0x37, 0x44, 0xdb, 0x03, 0x5d, 0xc6,
	0x1a, 0x0b, 0xad, 0x1a, 0x65, 0x1e, 0xe1, 0xc9, 0x76, 0x29, 0xbb, 0x89, 0xdb, 0x52, 0xf0, 0x11,
	0xdd, 0x34, 0xa1, 0x97, 0x80, 0x5c, 0x72, 0x2f, 0x8e, 0xc6, 0xe3, 0x7c, 0x24, 0xe3, 0xf4, 0x08,
	0xbd, 0x17, 0xa9, 0xee, 0x6e, 0x1b, 0xd1, 0x5b, 0x30, 0xe4, 0x35, 0x8a, 0x08, 0xbd, 0xf5, 0x9d,
	0x78, 0x0c, 0x45, 0xcb, 0x30, 0x88, 0x88, 0x01, 0x3c, 0xe2, 0xf3, 0x5c, 0x1b, 0x62, 0xdc, 0x4d,
	0x02, 0xc3, 0x02, 0x65, 0x29, 0xb8, 0xe1, 0xec, 0xf4, 0xf0, 0xb1, 0x44, 0x89, 0xcd, 0x42, 0x6a,
	0x14, 0xb9, 0xae, 0xe4, 0x72, 0xeb, 0xf2, 0x56, 0xb6, 0x2e, 0x6f, 0xa7, 0x24, 0x24, 0x67, 0xfe,
	0x02, 0x47, 0x5b, 0xf5, 0x40, 0x9f, 0x40, 0x25, 0xa3, 0x26, 0x83, 0x43, 0xc3, 0x90, 0x72, 0xa3,
	0xdc, 0x50, 0xd9, 0x7b, 0x60, 0xa8, 0x9c, 0x43, 0xfd, 0x5e, 0x95, 0x1e, 0x8c, 0x6f, 0xfe, 0xa9,
	0x42, 0x63, 0x77, 0x29, 0xe2, 0xa1, 0x93, 0x0c, 0x04, 0x31, 0x08, 0xc4, 0x2a, 0x2e, 0x44, 0xb4,
	0x72, 0x1c, 0x12, 0x25, 0x17, 0xac, 0x8c, 0xe5, 0x12, 0x8d, 0xe0, 0xc9, 0x8e, 0xf2, 0xaf, 0xe6,
	0x2c, 0x12, 0xa3, 0xe9, 0xd9, 0x7f, 0x56, 0x7f, 0x35, 0x67, 0xf8, 0xb1, 0xb3, 0xc3, 0x1a, 0x99,
	0xbf, 0xc3, 0xa3, 0x5d, 0xf0, 0x87, 0x66, 0xe6, 0x57, 0x50, 0xb7, 0x57, 0x6c, 0x16, 0x52, 0xff,
	0x37, 0xe2, 0x8e, 0x73, 0x45, 0xd4, 0xb0, 0x9e, 0x39, 0x78, 0x67, 0xa2, 0xcd, 0x31, 0xbf, 0xbf,
	0x39, 0xe6, 0xbf, 0x7c, 0x07, 0x5a, 0xfa, 0xb9, 0x43, 0x55, 0x28, 0xe3, 0xfe, 0x8f, 0x83, 0x9b,
	0x51, 0x1f, 0xeb, 0x0a, 0xd2, 0xe0, 0xa0, 0x73, 0x75, 0xdd, 0x7d, 0xa5, 0xab, 0xa8, 0x06, 0x5a,
	0xd7, 0xba, 0x18, 0x0c, 0xbb, 0xd7, 0xbd, 0xbe, 0xbe, 0x17, 0x2f, 0x71, 0xff, 0x65, 0xbf, 0x3b,
	0x1a, 0x5c, 0x0f, 0xf5, 0x7d, 0x54, 0x87, 0xda, 0xe5, 0xe0, 0x6a, 0xd4, 0xc7, 0xfd, 0x5e, 0x42,
	0x28, 0xa0, 0x63, 0x38, 0x92, 0x91, 0xba, 0xd6, 0xc5, 0x70, 0xd8, 0xbf, 0xd2, 0x0f, 0xda, 0xdf,
	0x41, 0x51, 0xa4, 0x74, 0x06, 0x85, 0xee, 0xcc, 0x66, 0x28, 0xfd, 0x34, 0xe5, 0xa6, 0xe0, 0x49,
	0x6d, 0xe3, 0x3b, 0x6c, 0x2a, 0xa7, 0xea, 0x99, 0xda, 0xf9, 0x15, 0xcc, 0x90, 0x7a, 0xcd, 0xd9,
	0xdd, 0x92, 0xd0, 0x39, 0x71, 0x3d, 0x42, 0x9b, 0x53, 0x7b, 0x42, 0x7d, 0x47, 0x82, 0xe3, 0xff,
	0x88, 0x4e, 0x2d, 0x89, 0xff, 0xc6, 0x76, 0xde, 0xdb, 0x1e, 0x79, 0xfb, 0x85, 0xe7, 0xb3, 0xd9,
	0x6a, 0x12, 0x8f, 0xb3, 0x56, 0x8e, 0xd9, 0x4a, 0x98, 0xad, 0x84, 0xd9, 0x8a, 0x99, 0x93, 0xe4,
	0x07, 0xe6, 0x9b, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0xad, 0x8f, 0x30, 0xf9, 0xdc, 0x08, 0x00,
	0x00,
}
