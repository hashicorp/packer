// Code generated by protoc-gen-go. DO NOT EDIT.
// source: yandex/cloud/loadbalancer/v1/health_check.proto

package loadbalancer // import "github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import duration "github.com/golang/protobuf/ptypes/duration"
import _ "github.com/yandex-cloud/go-genproto/yandex/cloud/validation"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// A HealthCheck resource. For more information, see [Health check](/docs/load-balancer/concepts/health-check).
type HealthCheck struct {
	// Name of the health check. The name must be unique for each target group that attached to a single load balancer. 3-63 characters long.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The interval between health checks. The default is 2 seconds.
	Interval *duration.Duration `protobuf:"bytes,2,opt,name=interval,proto3" json:"interval,omitempty"`
	// Timeout for a target to return a response for the health check. The default is 1 second.
	Timeout *duration.Duration `protobuf:"bytes,3,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// Number of failed health checks before changing the status to `` UNHEALTHY ``. The default is 2.
	UnhealthyThreshold int64 `protobuf:"varint,4,opt,name=unhealthy_threshold,json=unhealthyThreshold,proto3" json:"unhealthy_threshold,omitempty"`
	// Number of successful health checks required in order to set the `` HEALTHY `` status for the target. The default is 2.
	HealthyThreshold int64 `protobuf:"varint,5,opt,name=healthy_threshold,json=healthyThreshold,proto3" json:"healthy_threshold,omitempty"`
	// Protocol to use for the health check. Either TCP or HTTP.
	//
	// Types that are valid to be assigned to Options:
	//	*HealthCheck_TcpOptions_
	//	*HealthCheck_HttpOptions_
	Options              isHealthCheck_Options `protobuf_oneof:"options"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *HealthCheck) Reset()         { *m = HealthCheck{} }
func (m *HealthCheck) String() string { return proto.CompactTextString(m) }
func (*HealthCheck) ProtoMessage()    {}
func (*HealthCheck) Descriptor() ([]byte, []int) {
	return fileDescriptor_health_check_f79463e9bdeb651f, []int{0}
}
func (m *HealthCheck) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthCheck.Unmarshal(m, b)
}
func (m *HealthCheck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthCheck.Marshal(b, m, deterministic)
}
func (dst *HealthCheck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthCheck.Merge(dst, src)
}
func (m *HealthCheck) XXX_Size() int {
	return xxx_messageInfo_HealthCheck.Size(m)
}
func (m *HealthCheck) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthCheck.DiscardUnknown(m)
}

var xxx_messageInfo_HealthCheck proto.InternalMessageInfo

func (m *HealthCheck) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *HealthCheck) GetInterval() *duration.Duration {
	if m != nil {
		return m.Interval
	}
	return nil
}

func (m *HealthCheck) GetTimeout() *duration.Duration {
	if m != nil {
		return m.Timeout
	}
	return nil
}

func (m *HealthCheck) GetUnhealthyThreshold() int64 {
	if m != nil {
		return m.UnhealthyThreshold
	}
	return 0
}

func (m *HealthCheck) GetHealthyThreshold() int64 {
	if m != nil {
		return m.HealthyThreshold
	}
	return 0
}

type isHealthCheck_Options interface {
	isHealthCheck_Options()
}

type HealthCheck_TcpOptions_ struct {
	TcpOptions *HealthCheck_TcpOptions `protobuf:"bytes,6,opt,name=tcp_options,json=tcpOptions,proto3,oneof"`
}

type HealthCheck_HttpOptions_ struct {
	HttpOptions *HealthCheck_HttpOptions `protobuf:"bytes,7,opt,name=http_options,json=httpOptions,proto3,oneof"`
}

func (*HealthCheck_TcpOptions_) isHealthCheck_Options() {}

func (*HealthCheck_HttpOptions_) isHealthCheck_Options() {}

func (m *HealthCheck) GetOptions() isHealthCheck_Options {
	if m != nil {
		return m.Options
	}
	return nil
}

func (m *HealthCheck) GetTcpOptions() *HealthCheck_TcpOptions {
	if x, ok := m.GetOptions().(*HealthCheck_TcpOptions_); ok {
		return x.TcpOptions
	}
	return nil
}

func (m *HealthCheck) GetHttpOptions() *HealthCheck_HttpOptions {
	if x, ok := m.GetOptions().(*HealthCheck_HttpOptions_); ok {
		return x.HttpOptions
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*HealthCheck) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _HealthCheck_OneofMarshaler, _HealthCheck_OneofUnmarshaler, _HealthCheck_OneofSizer, []interface{}{
		(*HealthCheck_TcpOptions_)(nil),
		(*HealthCheck_HttpOptions_)(nil),
	}
}

func _HealthCheck_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*HealthCheck)
	// options
	switch x := m.Options.(type) {
	case *HealthCheck_TcpOptions_:
		b.EncodeVarint(6<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.TcpOptions); err != nil {
			return err
		}
	case *HealthCheck_HttpOptions_:
		b.EncodeVarint(7<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.HttpOptions); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("HealthCheck.Options has unexpected type %T", x)
	}
	return nil
}

func _HealthCheck_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*HealthCheck)
	switch tag {
	case 6: // options.tcp_options
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(HealthCheck_TcpOptions)
		err := b.DecodeMessage(msg)
		m.Options = &HealthCheck_TcpOptions_{msg}
		return true, err
	case 7: // options.http_options
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(HealthCheck_HttpOptions)
		err := b.DecodeMessage(msg)
		m.Options = &HealthCheck_HttpOptions_{msg}
		return true, err
	default:
		return false, nil
	}
}

func _HealthCheck_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*HealthCheck)
	// options
	switch x := m.Options.(type) {
	case *HealthCheck_TcpOptions_:
		s := proto.Size(x.TcpOptions)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *HealthCheck_HttpOptions_:
		s := proto.Size(x.HttpOptions)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Configuration option for a TCP health check.
type HealthCheck_TcpOptions struct {
	// Port to use for TCP health checks.
	Port                 int64    `protobuf:"varint,1,opt,name=port,proto3" json:"port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HealthCheck_TcpOptions) Reset()         { *m = HealthCheck_TcpOptions{} }
func (m *HealthCheck_TcpOptions) String() string { return proto.CompactTextString(m) }
func (*HealthCheck_TcpOptions) ProtoMessage()    {}
func (*HealthCheck_TcpOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_health_check_f79463e9bdeb651f, []int{0, 0}
}
func (m *HealthCheck_TcpOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthCheck_TcpOptions.Unmarshal(m, b)
}
func (m *HealthCheck_TcpOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthCheck_TcpOptions.Marshal(b, m, deterministic)
}
func (dst *HealthCheck_TcpOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthCheck_TcpOptions.Merge(dst, src)
}
func (m *HealthCheck_TcpOptions) XXX_Size() int {
	return xxx_messageInfo_HealthCheck_TcpOptions.Size(m)
}
func (m *HealthCheck_TcpOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthCheck_TcpOptions.DiscardUnknown(m)
}

var xxx_messageInfo_HealthCheck_TcpOptions proto.InternalMessageInfo

func (m *HealthCheck_TcpOptions) GetPort() int64 {
	if m != nil {
		return m.Port
	}
	return 0
}

// Configuration option for an HTTP health check.
type HealthCheck_HttpOptions struct {
	// Port to use for HTTP health checks.
	Port int64 `protobuf:"varint,1,opt,name=port,proto3" json:"port,omitempty"`
	// URL path to set for health checking requests for every target in the target group.
	// For example `` /ping ``. The default path is `` / ``.
	Path                 string   `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HealthCheck_HttpOptions) Reset()         { *m = HealthCheck_HttpOptions{} }
func (m *HealthCheck_HttpOptions) String() string { return proto.CompactTextString(m) }
func (*HealthCheck_HttpOptions) ProtoMessage()    {}
func (*HealthCheck_HttpOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_health_check_f79463e9bdeb651f, []int{0, 1}
}
func (m *HealthCheck_HttpOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthCheck_HttpOptions.Unmarshal(m, b)
}
func (m *HealthCheck_HttpOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthCheck_HttpOptions.Marshal(b, m, deterministic)
}
func (dst *HealthCheck_HttpOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthCheck_HttpOptions.Merge(dst, src)
}
func (m *HealthCheck_HttpOptions) XXX_Size() int {
	return xxx_messageInfo_HealthCheck_HttpOptions.Size(m)
}
func (m *HealthCheck_HttpOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthCheck_HttpOptions.DiscardUnknown(m)
}

var xxx_messageInfo_HealthCheck_HttpOptions proto.InternalMessageInfo

func (m *HealthCheck_HttpOptions) GetPort() int64 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *HealthCheck_HttpOptions) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func init() {
	proto.RegisterType((*HealthCheck)(nil), "yandex.cloud.loadbalancer.v1.HealthCheck")
	proto.RegisterType((*HealthCheck_TcpOptions)(nil), "yandex.cloud.loadbalancer.v1.HealthCheck.TcpOptions")
	proto.RegisterType((*HealthCheck_HttpOptions)(nil), "yandex.cloud.loadbalancer.v1.HealthCheck.HttpOptions")
}

func init() {
	proto.RegisterFile("yandex/cloud/loadbalancer/v1/health_check.proto", fileDescriptor_health_check_f79463e9bdeb651f)
}

var fileDescriptor_health_check_f79463e9bdeb651f = []byte{
	// 441 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x93, 0xcf, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0x31, 0x31, 0x4d, 0x3b, 0x46, 0x02, 0x96, 0x8b, 0x89, 0x28, 0x04, 0x4e, 0x39, 0xb0,
	0xeb, 0x6e, 0x42, 0x5a, 0x55, 0xdc, 0x0c, 0x87, 0x5c, 0x50, 0x25, 0xab, 0x12, 0x52, 0xa3, 0x2a,
	0xda, 0xd8, 0x8b, 0xd7, 0x62, 0xe3, 0xb5, 0xdc, 0x71, 0x44, 0x0b, 0xbc, 0x1b, 0x9c, 0xf2, 0x10,
	0xbc, 0x05, 0xc7, 0x9e, 0x50, 0xd7, 0xf9, 0x63, 0xa8, 0xd4, 0xf6, 0xe6, 0xd1, 0x7e, 0xbf, 0x6f,
	0x3e, 0xef, 0xcc, 0x42, 0x70, 0x2e, 0xf2, 0x44, 0x7e, 0x0d, 0x62, 0x6d, 0xaa, 0x24, 0xd0, 0x46,
	0x24, 0x53, 0xa1, 0x45, 0x1e, 0xcb, 0x32, 0x98, 0xf3, 0x40, 0x49, 0xa1, 0x51, 0x4d, 0x62, 0x25,
	0xe3, 0x2f, 0xac, 0x28, 0x0d, 0x1a, 0xf2, 0xbc, 0x06, 0x98, 0x05, 0x58, 0x13, 0x60, 0x73, 0xde,
	0x79, 0x91, 0x1a, 0x93, 0x6a, 0x19, 0x58, 0xed, 0xb4, 0xfa, 0x1c, 0x24, 0x55, 0x29, 0x30, 0x33,
	0x79, 0x4d, 0x77, 0x76, 0xff, 0x69, 0x37, 0x17, 0x3a, 0x4b, 0x1a, 0xc7, 0xaf, 0x7f, 0xbb, 0xe0,
	0x8d, 0x6c, 0xcf, 0xf7, 0x57, 0x2d, 0xc9, 0x10, 0xdc, 0x5c, 0xcc, 0xa4, 0xef, 0x74, 0x9d, 0xde,
	0x4e, 0xf8, 0xea, 0xcf, 0x82, 0xef, 0x7e, 0x1f, 0x0b, 0x7a, 0x71, 0x3a, 0xa6, 0x82, 0x5e, 0xec,
	0xd1, 0xc3, 0xd3, 0x6f, 0xfc, 0xcd, 0x3e, 0xff, 0x31, 0x5e, 0x56, 0x91, 0x95, 0x93, 0x21, 0x6c,
	0x67, 0x39, 0xca, 0x72, 0x2e, 0xb4, 0x7f, 0xbf, 0xeb, 0xf4, 0xbc, 0xfe, 0x33, 0x56, 0x07, 0x63,
	0xab, 0x60, 0xec, 0xc3, 0x32, 0x58, 0xb4, 0x96, 0x92, 0x01, 0xb4, 0x31, 0x9b, 0x49, 0x53, 0xa1,
	0xdf, 0xba, 0x8d, 0x5a, 0x29, 0xc9, 0x21, 0x3c, 0xad, 0xf2, 0xfa, 0x9e, 0xce, 0x27, 0xa8, 0x4a,
	0x79, 0xa6, 0x8c, 0x4e, 0x7c, 0xb7, 0xeb, 0xf4, 0x5a, 0xe1, 0xf6, 0xe5, 0x82, 0xbb, 0x7d, 0xca,
	0xf7, 0x22, 0xb2, 0x16, 0x1d, 0xaf, 0x34, 0x64, 0x08, 0x4f, 0xae, 0x83, 0x0f, 0xfe, 0x03, 0x1f,
	0x5f, 0xc3, 0x3e, 0x81, 0x87, 0x71, 0x31, 0x31, 0xc5, 0x55, 0x90, 0x33, 0x7f, 0xcb, 0x46, 0x7d,
	0xcb, 0x6e, 0x9a, 0x0b, 0x6b, 0x5c, 0x2a, 0x3b, 0x8e, 0x8b, 0xa3, 0x9a, 0x1d, 0xdd, 0x8b, 0x00,
	0xd7, 0x15, 0x39, 0x81, 0x87, 0x0a, 0x71, 0xe3, 0xdc, 0xb6, 0xce, 0xc3, 0xbb, 0x3b, 0x8f, 0x10,
	0x1b, 0xd6, 0x9e, 0xda, 0x94, 0x1d, 0x0a, 0xb0, 0xe9, 0x4b, 0x5e, 0x82, 0x5b, 0x98, 0x12, 0xed,
	0x5c, 0x5b, 0xa1, 0x77, 0xb9, 0xe0, 0x6d, 0x4e, 0x07, 0xfd, 0x83, 0xfd, 0x83, 0xc8, 0x1e, 0x74,
	0x42, 0xf0, 0x1a, 0x66, 0xb7, 0xea, 0x09, 0x01, 0xb7, 0x10, 0xa8, 0xec, 0xb4, 0x77, 0x22, 0xfb,
	0x1d, 0x3e, 0x82, 0xf6, 0xf2, 0x4f, 0x88, 0xfb, 0xf3, 0x17, 0x77, 0xc2, 0xa3, 0x93, 0x8f, 0x69,
	0x86, 0xaa, 0x9a, 0xb2, 0xd8, 0xcc, 0x96, 0x8b, 0x4f, 0xeb, 0x4d, 0x4c, 0x0d, 0x4d, 0x65, 0x6e,
	0xc7, 0x7c, 0xe3, 0x8b, 0x78, 0xd7, 0xac, 0xa7, 0x5b, 0x16, 0x18, 0xfc, 0x0d, 0x00, 0x00, 0xff,
	0xff, 0xaf, 0x9c, 0x24, 0xf4, 0x45, 0x03, 0x00, 0x00,
}
