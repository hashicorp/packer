// Code generated by protoc-gen-go. DO NOT EDIT.
// source: yandex/cloud/mdb/redis/v1/config/redis5_0.proto

package redis // import "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import wrappers "github.com/golang/protobuf/ptypes/wrappers"
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

type RedisConfig5_0_MaxmemoryPolicy int32

const (
	RedisConfig5_0_MAXMEMORY_POLICY_UNSPECIFIED RedisConfig5_0_MaxmemoryPolicy = 0
	// Try to remove less recently used (LRU) keys with `expire set`.
	RedisConfig5_0_VOLATILE_LRU RedisConfig5_0_MaxmemoryPolicy = 1
	// Remove less recently used (LRU) keys.
	RedisConfig5_0_ALLKEYS_LRU RedisConfig5_0_MaxmemoryPolicy = 2
	// Try to remove least frequently used (LFU) keys with `expire set`.
	RedisConfig5_0_VOLATILE_LFU RedisConfig5_0_MaxmemoryPolicy = 3
	// Remove least frequently used (LFU) keys.
	RedisConfig5_0_ALLKEYS_LFU RedisConfig5_0_MaxmemoryPolicy = 4
	// Try to remove keys with `expire set` randomly.
	RedisConfig5_0_VOLATILE_RANDOM RedisConfig5_0_MaxmemoryPolicy = 5
	// Remove keys randomly.
	RedisConfig5_0_ALLKEYS_RANDOM RedisConfig5_0_MaxmemoryPolicy = 6
	// Try to remove less recently used (LRU) keys with `expire set`
	// and shorter TTL first.
	RedisConfig5_0_VOLATILE_TTL RedisConfig5_0_MaxmemoryPolicy = 7
	// Return errors when memory limit was reached and commands could require
	// more memory to be used.
	RedisConfig5_0_NOEVICTION RedisConfig5_0_MaxmemoryPolicy = 8
)

var RedisConfig5_0_MaxmemoryPolicy_name = map[int32]string{
	0: "MAXMEMORY_POLICY_UNSPECIFIED",
	1: "VOLATILE_LRU",
	2: "ALLKEYS_LRU",
	3: "VOLATILE_LFU",
	4: "ALLKEYS_LFU",
	5: "VOLATILE_RANDOM",
	6: "ALLKEYS_RANDOM",
	7: "VOLATILE_TTL",
	8: "NOEVICTION",
}
var RedisConfig5_0_MaxmemoryPolicy_value = map[string]int32{
	"MAXMEMORY_POLICY_UNSPECIFIED": 0,
	"VOLATILE_LRU":                 1,
	"ALLKEYS_LRU":                  2,
	"VOLATILE_LFU":                 3,
	"ALLKEYS_LFU":                  4,
	"VOLATILE_RANDOM":              5,
	"ALLKEYS_RANDOM":               6,
	"VOLATILE_TTL":                 7,
	"NOEVICTION":                   8,
}

func (x RedisConfig5_0_MaxmemoryPolicy) String() string {
	return proto.EnumName(RedisConfig5_0_MaxmemoryPolicy_name, int32(x))
}
func (RedisConfig5_0_MaxmemoryPolicy) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_redis5_0_34508bcc9ea4303f, []int{0, 0}
}

// Fields and structure of `RedisConfig` reflects Redis configuration file
// parameters.
type RedisConfig5_0 struct {
	// Redis key eviction policy for a dataset that reaches maximum memory,
	// available to the host. Redis maxmemory setting depends on Managed
	// Service for Redis [host class](/docs/managed-redis/concepts/instance-types).
	//
	// All policies are described in detail in [Redis documentation](https://redis.io/topics/lru-cache).
	MaxmemoryPolicy RedisConfig5_0_MaxmemoryPolicy `protobuf:"varint,1,opt,name=maxmemory_policy,json=maxmemoryPolicy,proto3,enum=yandex.cloud.mdb.redis.v1.config.RedisConfig5_0_MaxmemoryPolicy" json:"maxmemory_policy,omitempty"`
	// Time that Redis keeps the connection open while the client is idle.
	// If no new command is sent during that time, the connection is closed.
	Timeout *wrappers.Int64Value `protobuf:"bytes,2,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// Authentication password.
	Password             string   `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RedisConfig5_0) Reset()         { *m = RedisConfig5_0{} }
func (m *RedisConfig5_0) String() string { return proto.CompactTextString(m) }
func (*RedisConfig5_0) ProtoMessage()    {}
func (*RedisConfig5_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_redis5_0_34508bcc9ea4303f, []int{0}
}
func (m *RedisConfig5_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RedisConfig5_0.Unmarshal(m, b)
}
func (m *RedisConfig5_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RedisConfig5_0.Marshal(b, m, deterministic)
}
func (dst *RedisConfig5_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RedisConfig5_0.Merge(dst, src)
}
func (m *RedisConfig5_0) XXX_Size() int {
	return xxx_messageInfo_RedisConfig5_0.Size(m)
}
func (m *RedisConfig5_0) XXX_DiscardUnknown() {
	xxx_messageInfo_RedisConfig5_0.DiscardUnknown(m)
}

var xxx_messageInfo_RedisConfig5_0 proto.InternalMessageInfo

func (m *RedisConfig5_0) GetMaxmemoryPolicy() RedisConfig5_0_MaxmemoryPolicy {
	if m != nil {
		return m.MaxmemoryPolicy
	}
	return RedisConfig5_0_MAXMEMORY_POLICY_UNSPECIFIED
}

func (m *RedisConfig5_0) GetTimeout() *wrappers.Int64Value {
	if m != nil {
		return m.Timeout
	}
	return nil
}

func (m *RedisConfig5_0) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type RedisConfigSet5_0 struct {
	// Effective settings for a Redis 5.0 cluster (a combination of settings
	// defined in [user_config] and [default_config]).
	EffectiveConfig *RedisConfig5_0 `protobuf:"bytes,1,opt,name=effective_config,json=effectiveConfig,proto3" json:"effective_config,omitempty"`
	// User-defined settings for a Redis 5.0 cluster.
	UserConfig *RedisConfig5_0 `protobuf:"bytes,2,opt,name=user_config,json=userConfig,proto3" json:"user_config,omitempty"`
	// Default configuration for a Redis 5.0 cluster.
	DefaultConfig        *RedisConfig5_0 `protobuf:"bytes,3,opt,name=default_config,json=defaultConfig,proto3" json:"default_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *RedisConfigSet5_0) Reset()         { *m = RedisConfigSet5_0{} }
func (m *RedisConfigSet5_0) String() string { return proto.CompactTextString(m) }
func (*RedisConfigSet5_0) ProtoMessage()    {}
func (*RedisConfigSet5_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_redis5_0_34508bcc9ea4303f, []int{1}
}
func (m *RedisConfigSet5_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RedisConfigSet5_0.Unmarshal(m, b)
}
func (m *RedisConfigSet5_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RedisConfigSet5_0.Marshal(b, m, deterministic)
}
func (dst *RedisConfigSet5_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RedisConfigSet5_0.Merge(dst, src)
}
func (m *RedisConfigSet5_0) XXX_Size() int {
	return xxx_messageInfo_RedisConfigSet5_0.Size(m)
}
func (m *RedisConfigSet5_0) XXX_DiscardUnknown() {
	xxx_messageInfo_RedisConfigSet5_0.DiscardUnknown(m)
}

var xxx_messageInfo_RedisConfigSet5_0 proto.InternalMessageInfo

func (m *RedisConfigSet5_0) GetEffectiveConfig() *RedisConfig5_0 {
	if m != nil {
		return m.EffectiveConfig
	}
	return nil
}

func (m *RedisConfigSet5_0) GetUserConfig() *RedisConfig5_0 {
	if m != nil {
		return m.UserConfig
	}
	return nil
}

func (m *RedisConfigSet5_0) GetDefaultConfig() *RedisConfig5_0 {
	if m != nil {
		return m.DefaultConfig
	}
	return nil
}

func init() {
	proto.RegisterType((*RedisConfig5_0)(nil), "yandex.cloud.mdb.redis.v1.config.RedisConfig5_0")
	proto.RegisterType((*RedisConfigSet5_0)(nil), "yandex.cloud.mdb.redis.v1.config.RedisConfigSet5_0")
	proto.RegisterEnum("yandex.cloud.mdb.redis.v1.config.RedisConfig5_0_MaxmemoryPolicy", RedisConfig5_0_MaxmemoryPolicy_name, RedisConfig5_0_MaxmemoryPolicy_value)
}

func init() {
	proto.RegisterFile("yandex/cloud/mdb/redis/v1/config/redis5_0.proto", fileDescriptor_redis5_0_34508bcc9ea4303f)
}

var fileDescriptor_redis5_0_34508bcc9ea4303f = []byte{
	// 528 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xdd, 0x8e, 0xd2, 0x4c,
	0x18, 0xc7, 0xdf, 0xc2, 0xeb, 0xee, 0x3a, 0x28, 0x8c, 0xe3, 0x09, 0x59, 0x3f, 0x82, 0x68, 0x36,
	0x44, 0xb7, 0x53, 0x40, 0x31, 0x6b, 0xfc, 0x5a, 0x60, 0x4b, 0x52, 0x2d, 0x14, 0xcb, 0x87, 0xb2,
	0x1b, 0x6d, 0x0a, 0x1d, 0x6a, 0x63, 0xcb, 0x90, 0x7e, 0xb0, 0x8b, 0xc6, 0x3b, 0xf1, 0x5e, 0x3c,
	0x31, 0xf1, 0x7e, 0xbc, 0x02, 0xc3, 0x0c, 0x25, 0xd6, 0x93, 0x8d, 0x7b, 0x38, 0xff, 0xfe, 0x9e,
	0x5f, 0xe7, 0xe9, 0xd3, 0x07, 0x48, 0x4b, 0x73, 0x66, 0x91, 0x33, 0x69, 0xe2, 0xd2, 0xc8, 0x92,
	0x3c, 0x6b, 0x2c, 0xf9, 0xc4, 0x72, 0x02, 0x69, 0x51, 0x91, 0x26, 0x74, 0x36, 0x75, 0x6c, 0x7e,
	0xae, 0x19, 0x65, 0x3c, 0xf7, 0x69, 0x48, 0x51, 0x81, 0x17, 0x60, 0x56, 0x80, 0x3d, 0x6b, 0x8c,
	0x19, 0x80, 0x17, 0x15, 0xcc, 0x0b, 0x76, 0x6f, 0xdb, 0x94, 0xda, 0x2e, 0x91, 0x18, 0x3f, 0x8e,
	0xa6, 0xd2, 0xa9, 0x6f, 0xce, 0xe7, 0xc4, 0x0f, 0xb8, 0x61, 0xf7, 0x56, 0xe2, 0x95, 0x0b, 0xd3,
	0x75, 0x2c, 0x33, 0x74, 0xe8, 0x8c, 0x3f, 0x2e, 0x7e, 0x4f, 0x83, 0xac, 0xbe, 0x52, 0x36, 0x99,
	0xae, 0x66, 0x94, 0xd1, 0x27, 0x00, 0x3d, 0xf3, 0xcc, 0x23, 0x1e, 0xf5, 0x97, 0xc6, 0x9c, 0xba,
	0xce, 0x64, 0x99, 0x17, 0x0a, 0x42, 0x29, 0x5b, 0x3d, 0xc4, 0xe7, 0x5d, 0x07, 0x27, 0x5d, 0xb8,
	0x1d, 0x8b, 0xba, 0xcc, 0xa3, 0xe7, 0xbc, 0x64, 0x80, 0x6a, 0x60, 0x3b, 0x74, 0x3c, 0x42, 0xa3,
	0x30, 0x9f, 0x2a, 0x08, 0xa5, 0x4c, 0xf5, 0x06, 0xe6, 0x0d, 0xe1, 0xb8, 0x21, 0xac, 0xcc, 0xc2,
	0xc7, 0x8f, 0x86, 0xa6, 0x1b, 0x11, 0x3d, 0x66, 0x51, 0x03, 0xec, 0xcc, 0xcd, 0x20, 0x38, 0xa5,
	0xbe, 0x95, 0x4f, 0x17, 0x84, 0xd2, 0xe5, 0xc6, 0xde, 0xaf, 0x9f, 0x95, 0xe2, 0x89, 0x29, 0x7e,
	0xae, 0x8b, 0xc7, 0x65, 0xf1, 0xc9, 0xe1, 0xf3, 0x07, 0x2f, 0xef, 0xe3, 0xfd, 0x3b, 0x7b, 0x77,
	0xef, 0x7d, 0x78, 0xf6, 0xc2, 0x10, 0xdf, 0x7f, 0x39, 0xd8, 0xaf, 0x54, 0x0f, 0xbe, 0xea, 0x9b,
	0xba, 0xe2, 0x0f, 0x01, 0xe4, 0xfe, 0xba, 0x1f, 0x2a, 0x80, 0x9b, 0xed, 0xfa, 0xbb, 0xb6, 0xdc,
	0xd6, 0xf4, 0x91, 0xd1, 0xd5, 0x54, 0xa5, 0x39, 0x32, 0x06, 0x9d, 0x5e, 0x57, 0x6e, 0x2a, 0x2d,
	0x45, 0x3e, 0x82, 0xff, 0x21, 0x08, 0xae, 0x0c, 0x35, 0xb5, 0xde, 0x57, 0x54, 0xd9, 0x50, 0xf5,
	0x01, 0x14, 0x50, 0x0e, 0x64, 0xea, 0xaa, 0xfa, 0x5a, 0x1e, 0xf5, 0x58, 0x90, 0x4a, 0x22, 0xad,
	0x01, 0x4c, 0x27, 0x90, 0xd6, 0x00, 0xfe, 0x8f, 0xae, 0x83, 0xdc, 0x06, 0xd1, 0xeb, 0x9d, 0x23,
	0xad, 0x0d, 0x2f, 0x21, 0x04, 0xb2, 0x31, 0xb5, 0xce, 0xb6, 0x12, 0xae, 0x7e, 0x5f, 0x85, 0xdb,
	0x28, 0x0b, 0x40, 0x47, 0x93, 0x87, 0x4a, 0xb3, 0xaf, 0x68, 0x1d, 0xb8, 0x53, 0xfc, 0x96, 0x02,
	0xd7, 0xfe, 0xf8, 0xea, 0x3d, 0x12, 0xae, 0x86, 0x78, 0x02, 0x20, 0x99, 0x4e, 0xc9, 0x24, 0x74,
	0x16, 0xc4, 0xe0, 0xb3, 0x61, 0x43, 0xcc, 0x54, 0xcb, 0xff, 0x3a, 0x44, 0x3d, 0xb7, 0x31, 0xf1,
	0x0c, 0xbd, 0x01, 0x99, 0x28, 0x20, 0x7e, 0xec, 0x4d, 0x5d, 0xd0, 0x0b, 0x56, 0x92, 0xb5, 0xf2,
	0x2d, 0xc8, 0x5a, 0x64, 0x6a, 0x46, 0x6e, 0x18, 0x5b, 0xd3, 0x17, 0xb4, 0x5e, 0x5d, 0x7b, 0x78,
	0xd2, 0x50, 0x8f, 0x5f, 0xd9, 0x4e, 0xf8, 0x31, 0x1a, 0xe3, 0x09, 0xf5, 0xd6, 0xfb, 0x27, 0xf2,
	0x65, 0xb0, 0xa9, 0x68, 0x93, 0x19, 0xfb, 0xcf, 0xce, 0x5d, 0xcc, 0xa7, 0xec, 0x3c, 0xde, 0x62,
	0xf4, 0xc3, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x0a, 0x11, 0xaa, 0xbd, 0xc9, 0x03, 0x00, 0x00,
}
