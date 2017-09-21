// Code generated by protoc-gen-go. DO NOT EDIT.
// source: msp/msp_principal.proto

package msp

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type MSPPrincipal_Classification int32

const (
	MSPPrincipal_ROLE MSPPrincipal_Classification = 0
	// one of a member of MSP network, and the one of an
	// administrator of an MSP network
	MSPPrincipal_ORGANIZATION_UNIT MSPPrincipal_Classification = 1
	// groupping of entities, per MSP affiliation
	// E.g., this can well be represented by an MSP's
	// Organization unit
	MSPPrincipal_IDENTITY MSPPrincipal_Classification = 2
)

var MSPPrincipal_Classification_name = map[int32]string{
	0: "ROLE",
	1: "ORGANIZATION_UNIT",
	2: "IDENTITY",
}
var MSPPrincipal_Classification_value = map[string]int32{
	"ROLE":              0,
	"ORGANIZATION_UNIT": 1,
	"IDENTITY":          2,
}

func (x MSPPrincipal_Classification) String() string {
	return proto.EnumName(MSPPrincipal_Classification_name, int32(x))
}
func (MSPPrincipal_Classification) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor2, []int{0, 0}
}

type MSPRole_MSPRoleType int32

const (
	MSPRole_MEMBER MSPRole_MSPRoleType = 0
	MSPRole_ADMIN  MSPRole_MSPRoleType = 1
)

var MSPRole_MSPRoleType_name = map[int32]string{
	0: "MEMBER",
	1: "ADMIN",
}
var MSPRole_MSPRoleType_value = map[string]int32{
	"MEMBER": 0,
	"ADMIN":  1,
}

func (x MSPRole_MSPRoleType) String() string {
	return proto.EnumName(MSPRole_MSPRoleType_name, int32(x))
}
func (MSPRole_MSPRoleType) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{2, 0} }

// MSPPrincipal aims to represent an MSP-centric set of identities.
// In particular, this structure allows for definition of
//  - a group of identities that are member of the same MSP
//  - a group of identities that are member of the same organization unit
//    in the same MSP
//  - a group of identities that are administering a specific MSP
//  - a specific identity
// Expressing these groups is done given two fields of the fields below
//  - Classification, that defines the type of classification of identities
//    in an MSP this principal would be defined on; Classification can take
//    three values:
//     (i)  ByMSPRole: that represents a classification of identities within
//          MSP based on one of the two pre-defined MSP rules, "member" and "admin"
//     (ii) ByOrganizationUnit: that represents a classification of identities
//          within MSP based on the organization unit an identity belongs to
//     (iii)ByIdentity that denotes that MSPPrincipal is mapped to a single
//          identity/certificate; this would mean that the Principal bytes
//          message
type MSPPrincipal struct {
	// Classification describes the way that one should process
	// Principal. An Classification value of "ByOrganizationUnit" reflects
	// that "Principal" contains the name of an organization this MSP
	// handles. A Classification value "ByIdentity" means that
	// "Principal" contains a specific identity. Default value
	// denotes that Principal contains one of the groups by
	// default supported by all MSPs ("admin" or "member").
	PrincipalClassification MSPPrincipal_Classification `protobuf:"varint,1,opt,name=principal_classification,json=principalClassification,enum=common.MSPPrincipal_Classification" json:"principal_classification,omitempty"`
	// Principal completes the policy principal definition. For the default
	// principal types, Principal can be either "Admin" or "Member".
	// For the ByOrganizationUnit/ByIdentity values of Classification,
	// PolicyPrincipal acquires its value from an organization unit or
	// identity, respectively.
	Principal []byte `protobuf:"bytes,2,opt,name=principal,proto3" json:"principal,omitempty"`
}

func (m *MSPPrincipal) Reset()                    { *m = MSPPrincipal{} }
func (m *MSPPrincipal) String() string            { return proto.CompactTextString(m) }
func (*MSPPrincipal) ProtoMessage()               {}
func (*MSPPrincipal) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *MSPPrincipal) GetPrincipalClassification() MSPPrincipal_Classification {
	if m != nil {
		return m.PrincipalClassification
	}
	return MSPPrincipal_ROLE
}

func (m *MSPPrincipal) GetPrincipal() []byte {
	if m != nil {
		return m.Principal
	}
	return nil
}

// OrganizationUnit governs the organization of the Principal
// field of a policy principal when a specific organization unity members
// are to be defined within a policy principal.
type OrganizationUnit struct {
	// MSPIdentifier represents the identifier of the MSP this organization unit
	// refers to
	MspIdentifier string `protobuf:"bytes,1,opt,name=msp_identifier,json=mspIdentifier" json:"msp_identifier,omitempty"`
	// OrganizationUnitIdentifier defines the organizational unit under the
	// MSP identified with MSPIdentifier
	OrganizationalUnitIdentifier string `protobuf:"bytes,2,opt,name=organizational_unit_identifier,json=organizationalUnitIdentifier" json:"organizational_unit_identifier,omitempty"`
	// CertifiersIdentifier is the hash of certificates chain of trust
	// related to this organizational unit
	CertifiersIdentifier []byte `protobuf:"bytes,3,opt,name=certifiers_identifier,json=certifiersIdentifier,proto3" json:"certifiers_identifier,omitempty"`
}

func (m *OrganizationUnit) Reset()                    { *m = OrganizationUnit{} }
func (m *OrganizationUnit) String() string            { return proto.CompactTextString(m) }
func (*OrganizationUnit) ProtoMessage()               {}
func (*OrganizationUnit) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *OrganizationUnit) GetMspIdentifier() string {
	if m != nil {
		return m.MspIdentifier
	}
	return ""
}

func (m *OrganizationUnit) GetOrganizationalUnitIdentifier() string {
	if m != nil {
		return m.OrganizationalUnitIdentifier
	}
	return ""
}

func (m *OrganizationUnit) GetCertifiersIdentifier() []byte {
	if m != nil {
		return m.CertifiersIdentifier
	}
	return nil
}

// MSPRole governs the organization of the Principal
// field of an MSPPrincipal when it aims to define one of the
// two dedicated roles within an MSP: Admin and Members.
type MSPRole struct {
	// MSPIdentifier represents the identifier of the MSP this principal
	// refers to
	MspIdentifier string `protobuf:"bytes,1,opt,name=msp_identifier,json=mspIdentifier" json:"msp_identifier,omitempty"`
	// MSPRoleType defines which of the available, pre-defined MSP-roles
	// an identiy should posess inside the MSP with identifier MSPidentifier
	Role MSPRole_MSPRoleType `protobuf:"varint,2,opt,name=role,enum=common.MSPRole_MSPRoleType" json:"role,omitempty"`
}

func (m *MSPRole) Reset()                    { *m = MSPRole{} }
func (m *MSPRole) String() string            { return proto.CompactTextString(m) }
func (*MSPRole) ProtoMessage()               {}
func (*MSPRole) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{2} }

func (m *MSPRole) GetMspIdentifier() string {
	if m != nil {
		return m.MspIdentifier
	}
	return ""
}

func (m *MSPRole) GetRole() MSPRole_MSPRoleType {
	if m != nil {
		return m.Role
	}
	return MSPRole_MEMBER
}

func init() {
	proto.RegisterType((*MSPPrincipal)(nil), "common.MSPPrincipal")
	proto.RegisterType((*OrganizationUnit)(nil), "common.OrganizationUnit")
	proto.RegisterType((*MSPRole)(nil), "common.MSPRole")
	proto.RegisterEnum("common.MSPPrincipal_Classification", MSPPrincipal_Classification_name, MSPPrincipal_Classification_value)
	proto.RegisterEnum("common.MSPRole_MSPRoleType", MSPRole_MSPRoleType_name, MSPRole_MSPRoleType_value)
}

func init() { proto.RegisterFile("msp/msp_principal.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 386 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xdf, 0x8e, 0x93, 0x40,
	0x14, 0x87, 0x77, 0xea, 0x5a, 0xb7, 0xc7, 0x4a, 0x70, 0xe2, 0x66, 0x9b, 0xb8, 0x31, 0x1b, 0xac,
	0x49, 0xaf, 0x20, 0x69, 0x1f, 0xc0, 0xb4, 0x96, 0x18, 0x12, 0xf9, 0x93, 0x29, 0xbd, 0xb0, 0x17,
	0x12, 0x4a, 0xa7, 0x74, 0x12, 0x60, 0x26, 0x03, 0xbd, 0xa8, 0x2f, 0xe0, 0x0b, 0xf9, 0x1a, 0xbe,
	0x93, 0x01, 0x2c, 0x9d, 0x7a, 0xb5, 0x57, 0x30, 0xe7, 0xf7, 0x7d, 0x67, 0x66, 0xe0, 0xc0, 0x43,
	0x5e, 0x0a, 0x2b, 0x2f, 0x45, 0x24, 0x24, 0x2b, 0x12, 0x26, 0xe2, 0xcc, 0x14, 0x92, 0x57, 0x1c,
	0xf7, 0x13, 0x9e, 0xe7, 0xbc, 0x30, 0xfe, 0x20, 0x18, 0xba, 0xab, 0x20, 0x38, 0xc7, 0xf8, 0x07,
	0x8c, 0x3a, 0x36, 0x4a, 0xb2, 0xb8, 0x2c, 0xd9, 0x9e, 0x25, 0x71, 0xc5, 0x78, 0x31, 0x42, 0x4f,
	0x68, 0xa2, 0x4d, 0x3f, 0x9a, 0xad, 0x6b, 0xaa, 0x9e, 0xf9, 0xe5, 0x0a, 0x25, 0x0f, 0x5d, 0x93,
	0xeb, 0x00, 0x3f, 0xc2, 0xa0, 0x8b, 0x46, 0xbd, 0x27, 0x34, 0x19, 0x92, 0x4b, 0xc1, 0xf8, 0x0c,
	0xda, 0x7f, 0xfc, 0x1d, 0xdc, 0x12, 0xff, 0x9b, 0xad, 0xdf, 0xe0, 0x7b, 0x78, 0xeb, 0x93, 0xaf,
	0x73, 0xcf, 0xd9, 0xcc, 0x43, 0xc7, 0xf7, 0xa2, 0xb5, 0xe7, 0x84, 0x3a, 0xc2, 0x43, 0xb8, 0x73,
	0x96, 0xb6, 0x17, 0x3a, 0xe1, 0x77, 0xbd, 0x67, 0xfc, 0x46, 0xa0, 0xfb, 0x32, 0x8d, 0x0b, 0xf6,
	0xb3, 0xf1, 0xd7, 0x05, 0xab, 0xf0, 0x27, 0xd0, 0xea, 0x6f, 0xc0, 0x76, 0xb4, 0xa8, 0xd8, 0x9e,
	0x51, 0xd9, 0xdc, 0x64, 0x40, 0xde, 0xe4, 0xa5, 0x70, 0xba, 0x22, 0x5e, 0xc2, 0x07, 0xae, 0xa8,
	0x71, 0x16, 0x1d, 0x0b, 0x56, 0xa9, 0x5a, 0xaf, 0xd1, 0x1e, 0xaf, 0xa9, 0x7a, 0x0b, 0xa5, 0xcb,
	0x0c, 0xee, 0x13, 0x2a, 0xdb, 0x45, 0xa9, 0xca, 0x2f, 0x9a, 0xcb, 0xbe, 0xbb, 0x84, 0x17, 0xc9,
	0xf8, 0x85, 0xe0, 0x95, 0xbb, 0x0a, 0x08, 0xcf, 0xe8, 0x73, 0x4f, 0x6b, 0xc1, 0xad, 0xe4, 0x19,
	0x6d, 0xce, 0xa4, 0x4d, 0xdf, 0x2b, 0x3f, 0xa5, 0xee, 0x72, 0x7e, 0x86, 0x27, 0x41, 0x49, 0x03,
	0x1a, 0x63, 0x78, 0xad, 0x14, 0x31, 0x40, 0xdf, 0xb5, 0xdd, 0x85, 0x4d, 0xf4, 0x1b, 0x3c, 0x80,
	0x97, 0xf3, 0xa5, 0xeb, 0x78, 0x3a, 0x5a, 0x04, 0x30, 0xe6, 0x32, 0x35, 0x0f, 0x27, 0x41, 0x65,
	0x46, 0x77, 0x29, 0x95, 0xe6, 0x3e, 0xde, 0x4a, 0x96, 0xb4, 0x83, 0x53, 0xfe, 0xdb, 0x67, 0x33,
	0x49, 0x59, 0x75, 0x38, 0x6e, 0xeb, 0xa5, 0xa5, 0xc0, 0x56, 0x0b, 0x5b, 0x2d, 0x5c, 0x8f, 0xde,
	0xb6, 0xdf, 0xbc, 0xcf, 0xfe, 0x06, 0x00, 0x00, 0xff, 0xff, 0xf1, 0xd3, 0xfe, 0xdb, 0x8c, 0x02,
	0x00, 0x00,
}
