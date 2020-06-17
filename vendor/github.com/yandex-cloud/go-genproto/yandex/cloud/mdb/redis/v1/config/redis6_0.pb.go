// Code generated by protoc-gen-go. DO NOT EDIT.
// source: yandex/cloud/mdb/redis/v1/config/redis6_0.proto

package redis

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	_ "github.com/yandex-cloud/go-genproto/yandex/cloud"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type RedisConfig6_0_MaxmemoryPolicy int32

const (
	RedisConfig6_0_MAXMEMORY_POLICY_UNSPECIFIED RedisConfig6_0_MaxmemoryPolicy = 0
	// Try to remove less recently used (LRU) keys with `expire set`.
	RedisConfig6_0_VOLATILE_LRU RedisConfig6_0_MaxmemoryPolicy = 1
	// Remove less recently used (LRU) keys.
	RedisConfig6_0_ALLKEYS_LRU RedisConfig6_0_MaxmemoryPolicy = 2
	// Try to remove least frequently used (LFU) keys with `expire set`.
	RedisConfig6_0_VOLATILE_LFU RedisConfig6_0_MaxmemoryPolicy = 3
	// Remove least frequently used (LFU) keys.
	RedisConfig6_0_ALLKEYS_LFU RedisConfig6_0_MaxmemoryPolicy = 4
	// Try to remove keys with `expire set` randomly.
	RedisConfig6_0_VOLATILE_RANDOM RedisConfig6_0_MaxmemoryPolicy = 5
	// Remove keys randomly.
	RedisConfig6_0_ALLKEYS_RANDOM RedisConfig6_0_MaxmemoryPolicy = 6
	// Try to remove less recently used (LRU) keys with `expire set`
	// and shorter TTL first.
	RedisConfig6_0_VOLATILE_TTL RedisConfig6_0_MaxmemoryPolicy = 7
	// Return errors when memory limit was reached and commands could require
	// more memory to be used.
	RedisConfig6_0_NOEVICTION RedisConfig6_0_MaxmemoryPolicy = 8
)

var RedisConfig6_0_MaxmemoryPolicy_name = map[int32]string{
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

var RedisConfig6_0_MaxmemoryPolicy_value = map[string]int32{
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

func (x RedisConfig6_0_MaxmemoryPolicy) String() string {
	return proto.EnumName(RedisConfig6_0_MaxmemoryPolicy_name, int32(x))
}

func (RedisConfig6_0_MaxmemoryPolicy) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5861c944012bf0d1, []int{0, 0}
}

// Fields and structure of `RedisConfig` reflects Redis configuration file
// parameters.
type RedisConfig6_0 struct {
	// Redis key eviction policy for a dataset that reaches maximum memory,
	// available to the host. Redis maxmemory setting depends on Managed
	// Service for Redis [host class](/docs/managed-redis/concepts/instance-types).
	//
	// All policies are described in detail in [Redis documentation](https://redis.io/topics/lru-cache).
	MaxmemoryPolicy RedisConfig6_0_MaxmemoryPolicy `protobuf:"varint,1,opt,name=maxmemory_policy,json=maxmemoryPolicy,proto3,enum=yandex.cloud.mdb.redis.v1.config.RedisConfig6_0_MaxmemoryPolicy" json:"maxmemory_policy,omitempty"`
	// Time that Redis keeps the connection open while the client is idle.
	// If no new command is sent during that time, the connection is closed.
	Timeout *wrappers.Int64Value `protobuf:"bytes,2,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// Authentication password.
	Password             string   `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RedisConfig6_0) Reset()         { *m = RedisConfig6_0{} }
func (m *RedisConfig6_0) String() string { return proto.CompactTextString(m) }
func (*RedisConfig6_0) ProtoMessage()    {}
func (*RedisConfig6_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_5861c944012bf0d1, []int{0}
}

func (m *RedisConfig6_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RedisConfig6_0.Unmarshal(m, b)
}
func (m *RedisConfig6_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RedisConfig6_0.Marshal(b, m, deterministic)
}
func (m *RedisConfig6_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RedisConfig6_0.Merge(m, src)
}
func (m *RedisConfig6_0) XXX_Size() int {
	return xxx_messageInfo_RedisConfig6_0.Size(m)
}
func (m *RedisConfig6_0) XXX_DiscardUnknown() {
	xxx_messageInfo_RedisConfig6_0.DiscardUnknown(m)
}

var xxx_messageInfo_RedisConfig6_0 proto.InternalMessageInfo

func (m *RedisConfig6_0) GetMaxmemoryPolicy() RedisConfig6_0_MaxmemoryPolicy {
	if m != nil {
		return m.MaxmemoryPolicy
	}
	return RedisConfig6_0_MAXMEMORY_POLICY_UNSPECIFIED
}

func (m *RedisConfig6_0) GetTimeout() *wrappers.Int64Value {
	if m != nil {
		return m.Timeout
	}
	return nil
}

func (m *RedisConfig6_0) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

type RedisConfigSet6_0 struct {
	// Effective settings for a Redis 6.0 cluster (a combination of settings
	// defined in [user_config] and [default_config]).
	EffectiveConfig *RedisConfig6_0 `protobuf:"bytes,1,opt,name=effective_config,json=effectiveConfig,proto3" json:"effective_config,omitempty"`
	// User-defined settings for a Redis 6.0 cluster.
	UserConfig *RedisConfig6_0 `protobuf:"bytes,2,opt,name=user_config,json=userConfig,proto3" json:"user_config,omitempty"`
	// Default configuration for a Redis 6.0 cluster.
	DefaultConfig        *RedisConfig6_0 `protobuf:"bytes,3,opt,name=default_config,json=defaultConfig,proto3" json:"default_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *RedisConfigSet6_0) Reset()         { *m = RedisConfigSet6_0{} }
func (m *RedisConfigSet6_0) String() string { return proto.CompactTextString(m) }
func (*RedisConfigSet6_0) ProtoMessage()    {}
func (*RedisConfigSet6_0) Descriptor() ([]byte, []int) {
	return fileDescriptor_5861c944012bf0d1, []int{1}
}

func (m *RedisConfigSet6_0) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RedisConfigSet6_0.Unmarshal(m, b)
}
func (m *RedisConfigSet6_0) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RedisConfigSet6_0.Marshal(b, m, deterministic)
}
func (m *RedisConfigSet6_0) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RedisConfigSet6_0.Merge(m, src)
}
func (m *RedisConfigSet6_0) XXX_Size() int {
	return xxx_messageInfo_RedisConfigSet6_0.Size(m)
}
func (m *RedisConfigSet6_0) XXX_DiscardUnknown() {
	xxx_messageInfo_RedisConfigSet6_0.DiscardUnknown(m)
}

var xxx_messageInfo_RedisConfigSet6_0 proto.InternalMessageInfo

func (m *RedisConfigSet6_0) GetEffectiveConfig() *RedisConfig6_0 {
	if m != nil {
		return m.EffectiveConfig
	}
	return nil
}

func (m *RedisConfigSet6_0) GetUserConfig() *RedisConfig6_0 {
	if m != nil {
		return m.UserConfig
	}
	return nil
}

func (m *RedisConfigSet6_0) GetDefaultConfig() *RedisConfig6_0 {
	if m != nil {
		return m.DefaultConfig
	}
	return nil
}

func init() {
	proto.RegisterEnum("yandex.cloud.mdb.redis.v1.config.RedisConfig6_0_MaxmemoryPolicy", RedisConfig6_0_MaxmemoryPolicy_name, RedisConfig6_0_MaxmemoryPolicy_value)
	proto.RegisterType((*RedisConfig6_0)(nil), "yandex.cloud.mdb.redis.v1.config.RedisConfig6_0")
	proto.RegisterType((*RedisConfigSet6_0)(nil), "yandex.cloud.mdb.redis.v1.config.RedisConfigSet6_0")
}

func init() {
	proto.RegisterFile("yandex/cloud/mdb/redis/v1/config/redis6_0.proto", fileDescriptor_5861c944012bf0d1)
}

var fileDescriptor_5861c944012bf0d1 = []byte{
	// 535 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xed, 0x6e, 0xd2, 0x50,
	0x18, 0xc7, 0x2d, 0xe8, 0x36, 0x0f, 0x0a, 0xf5, 0xf8, 0x85, 0xcc, 0x97, 0x20, 0x2e, 0x0b, 0xd1,
	0xf5, 0x14, 0x50, 0xc9, 0x8c, 0x6f, 0x03, 0x56, 0x92, 0x6a, 0xa1, 0x58, 0x5e, 0x94, 0x2d, 0xda,
	0x14, 0x7a, 0xa8, 0x8d, 0x2d, 0xa7, 0xe9, 0x0b, 0x1b, 0x1a, 0xef, 0xc4, 0x7b, 0xf1, 0x8b, 0x89,
	0xf7, 0xe3, 0x15, 0x18, 0xce, 0xa1, 0xc4, 0x1a, 0x93, 0x65, 0xfb, 0x78, 0xfe, 0xfd, 0x3d, 0xbf,
	0x9e, 0xa7, 0x4f, 0x1f, 0x20, 0x2e, 0x8c, 0x99, 0x89, 0x4f, 0xc5, 0x89, 0x43, 0x22, 0x53, 0x74,
	0xcd, 0xb1, 0xe8, 0x63, 0xd3, 0x0e, 0xc4, 0x79, 0x45, 0x9c, 0x90, 0xd9, 0xd4, 0xb6, 0xd8, 0xb9,
	0xa6, 0x97, 0x91, 0xe7, 0x93, 0x90, 0xc0, 0x02, 0x2b, 0x40, 0xb4, 0x00, 0xb9, 0xe6, 0x18, 0x51,
	0x00, 0xcd, 0x2b, 0x88, 0x15, 0x6c, 0xdf, 0xb5, 0x08, 0xb1, 0x1c, 0x2c, 0x52, 0x7e, 0x1c, 0x4d,
	0xc5, 0x13, 0xdf, 0xf0, 0x3c, 0xec, 0x07, 0xcc, 0xb0, 0x7d, 0x27, 0xf1, 0xca, 0xb9, 0xe1, 0xd8,
	0xa6, 0x11, 0xda, 0x64, 0xc6, 0x1e, 0x17, 0x7f, 0xa4, 0x41, 0x56, 0x5b, 0x2a, 0x9b, 0x54, 0x57,
	0xd3, 0xcb, 0xf0, 0x33, 0xe0, 0x5d, 0xe3, 0xd4, 0xc5, 0x2e, 0xf1, 0x17, 0xba, 0x47, 0x1c, 0x7b,
	0xb2, 0xc8, 0x73, 0x05, 0xae, 0x94, 0xad, 0x1e, 0xa0, 0xb3, 0xae, 0x83, 0x92, 0x2e, 0xd4, 0x8e,
	0x45, 0x5d, 0xea, 0xd1, 0x72, 0x6e, 0x32, 0x80, 0x4f, 0xc0, 0x66, 0x68, 0xbb, 0x98, 0x44, 0x61,
	0x3e, 0x55, 0xe0, 0x4a, 0x99, 0xea, 0x2d, 0xc4, 0x1a, 0x42, 0x71, 0x43, 0x48, 0x9e, 0x85, 0xb5,
	0xc7, 0x43, 0xc3, 0x89, 0xb0, 0x16, 0xb3, 0xb0, 0x01, 0xb6, 0x3c, 0x23, 0x08, 0x4e, 0x88, 0x6f,
	0xe6, 0xd3, 0x05, 0xae, 0x74, 0xb5, 0xb1, 0xfb, 0xfb, 0x57, 0xa5, 0x78, 0x6c, 0x08, 0x5f, 0xea,
	0xc2, 0x51, 0x59, 0x78, 0x7a, 0xf0, 0xe2, 0xe1, 0xab, 0x07, 0x68, 0xef, 0xde, 0xee, 0xfd, 0x9d,
	0x8f, 0xcf, 0x5f, 0xea, 0xc2, 0x87, 0xaf, 0xfb, 0x7b, 0x95, 0xea, 0xfe, 0x37, 0x6d, 0x5d, 0x57,
	0xfc, 0xc9, 0x81, 0xdc, 0x3f, 0xf7, 0x83, 0x05, 0x70, 0xbb, 0x5d, 0x7f, 0xdf, 0x96, 0xda, 0xaa,
	0x36, 0xd2, 0xbb, 0xaa, 0x22, 0x37, 0x47, 0xfa, 0xa0, 0xd3, 0xeb, 0x4a, 0x4d, 0xb9, 0x25, 0x4b,
	0x87, 0xfc, 0x25, 0xc8, 0x83, 0x6b, 0x43, 0x55, 0xa9, 0xf7, 0x65, 0x45, 0xd2, 0x15, 0x6d, 0xc0,
	0x73, 0x30, 0x07, 0x32, 0x75, 0x45, 0x79, 0x23, 0x8d, 0x7a, 0x34, 0x48, 0x25, 0x91, 0xd6, 0x80,
	0x4f, 0x27, 0x90, 0xd6, 0x80, 0xbf, 0x0c, 0x6f, 0x82, 0xdc, 0x1a, 0xd1, 0xea, 0x9d, 0x43, 0xb5,
	0xcd, 0x5f, 0x81, 0x10, 0x64, 0x63, 0x6a, 0x95, 0x6d, 0x24, 0x5c, 0xfd, 0xbe, 0xc2, 0x6f, 0xc2,
	0x2c, 0x00, 0x1d, 0x55, 0x1a, 0xca, 0xcd, 0xbe, 0xac, 0x76, 0xf8, 0xad, 0xe2, 0xf7, 0x14, 0xb8,
	0xf1, 0xd7, 0x57, 0xef, 0xe1, 0x70, 0x39, 0xc4, 0x63, 0xc0, 0xe3, 0xe9, 0x14, 0x4f, 0x42, 0x7b,
	0x8e, 0x75, 0x36, 0x1b, 0x3a, 0xc4, 0x4c, 0xb5, 0x7c, 0xde, 0x21, 0x6a, 0xb9, 0xb5, 0x89, 0x65,
	0xf0, 0x2d, 0xc8, 0x44, 0x01, 0xf6, 0x63, 0x6f, 0xea, 0x82, 0x5e, 0xb0, 0x94, 0xac, 0x94, 0xef,
	0x40, 0xd6, 0xc4, 0x53, 0x23, 0x72, 0xc2, 0xd8, 0x9a, 0xbe, 0xa0, 0xf5, 0xfa, 0xca, 0xc3, 0x92,
	0x86, 0x0f, 0x76, 0x12, 0x06, 0xc3, 0xb3, 0xff, 0x67, 0x39, 0x7a, 0x6d, 0xd9, 0xe1, 0xa7, 0x68,
	0x8c, 0x26, 0xc4, 0x5d, 0x6d, 0xa9, 0xc0, 0x56, 0xc6, 0x22, 0x82, 0x85, 0x67, 0xf4, 0x6f, 0x3c,
	0x73, 0x7d, 0x9f, 0xd1, 0xf3, 0x78, 0x83, 0xd2, 0x8f, 0xfe, 0x04, 0x00, 0x00, 0xff, 0xff, 0xed,
	0x3e, 0x0c, 0x1f, 0xef, 0x03, 0x00, 0x00,
}
