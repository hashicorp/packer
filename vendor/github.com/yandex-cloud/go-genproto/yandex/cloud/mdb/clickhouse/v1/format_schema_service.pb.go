// Code generated by protoc-gen-go. DO NOT EDIT.
// source: yandex/cloud/mdb/clickhouse/v1/format_schema_service.proto

package clickhouse

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/yandex-cloud/go-genproto/yandex/cloud"
	_ "github.com/yandex-cloud/go-genproto/yandex/cloud/api"
	operation "github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	field_mask "google.golang.org/genproto/protobuf/field_mask"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type GetFormatSchemaRequest struct {
	ClusterId            string   `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string   `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetFormatSchemaRequest) Reset()         { *m = GetFormatSchemaRequest{} }
func (m *GetFormatSchemaRequest) String() string { return proto.CompactTextString(m) }
func (*GetFormatSchemaRequest) ProtoMessage()    {}
func (*GetFormatSchemaRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{0}
}

func (m *GetFormatSchemaRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetFormatSchemaRequest.Unmarshal(m, b)
}
func (m *GetFormatSchemaRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetFormatSchemaRequest.Marshal(b, m, deterministic)
}
func (m *GetFormatSchemaRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetFormatSchemaRequest.Merge(m, src)
}
func (m *GetFormatSchemaRequest) XXX_Size() int {
	return xxx_messageInfo_GetFormatSchemaRequest.Size(m)
}
func (m *GetFormatSchemaRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetFormatSchemaRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetFormatSchemaRequest proto.InternalMessageInfo

func (m *GetFormatSchemaRequest) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *GetFormatSchemaRequest) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

type ListFormatSchemasRequest struct {
	ClusterId            string   `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	PageSize             int64    `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	PageToken            string   `protobuf:"bytes,3,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListFormatSchemasRequest) Reset()         { *m = ListFormatSchemasRequest{} }
func (m *ListFormatSchemasRequest) String() string { return proto.CompactTextString(m) }
func (*ListFormatSchemasRequest) ProtoMessage()    {}
func (*ListFormatSchemasRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{1}
}

func (m *ListFormatSchemasRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListFormatSchemasRequest.Unmarshal(m, b)
}
func (m *ListFormatSchemasRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListFormatSchemasRequest.Marshal(b, m, deterministic)
}
func (m *ListFormatSchemasRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListFormatSchemasRequest.Merge(m, src)
}
func (m *ListFormatSchemasRequest) XXX_Size() int {
	return xxx_messageInfo_ListFormatSchemasRequest.Size(m)
}
func (m *ListFormatSchemasRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListFormatSchemasRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListFormatSchemasRequest proto.InternalMessageInfo

func (m *ListFormatSchemasRequest) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *ListFormatSchemasRequest) GetPageSize() int64 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

func (m *ListFormatSchemasRequest) GetPageToken() string {
	if m != nil {
		return m.PageToken
	}
	return ""
}

type ListFormatSchemasResponse struct {
	FormatSchemas        []*FormatSchema `protobuf:"bytes,1,rep,name=format_schemas,json=formatSchemas,proto3" json:"format_schemas,omitempty"`
	NextPageToken        string          `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ListFormatSchemasResponse) Reset()         { *m = ListFormatSchemasResponse{} }
func (m *ListFormatSchemasResponse) String() string { return proto.CompactTextString(m) }
func (*ListFormatSchemasResponse) ProtoMessage()    {}
func (*ListFormatSchemasResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{2}
}

func (m *ListFormatSchemasResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListFormatSchemasResponse.Unmarshal(m, b)
}
func (m *ListFormatSchemasResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListFormatSchemasResponse.Marshal(b, m, deterministic)
}
func (m *ListFormatSchemasResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListFormatSchemasResponse.Merge(m, src)
}
func (m *ListFormatSchemasResponse) XXX_Size() int {
	return xxx_messageInfo_ListFormatSchemasResponse.Size(m)
}
func (m *ListFormatSchemasResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListFormatSchemasResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListFormatSchemasResponse proto.InternalMessageInfo

func (m *ListFormatSchemasResponse) GetFormatSchemas() []*FormatSchema {
	if m != nil {
		return m.FormatSchemas
	}
	return nil
}

func (m *ListFormatSchemasResponse) GetNextPageToken() string {
	if m != nil {
		return m.NextPageToken
	}
	return ""
}

type CreateFormatSchemaRequest struct {
	ClusterId            string           `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string           `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	Type                 FormatSchemaType `protobuf:"varint,3,opt,name=type,proto3,enum=yandex.cloud.mdb.clickhouse.v1.FormatSchemaType" json:"type,omitempty"`
	Uri                  string           `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *CreateFormatSchemaRequest) Reset()         { *m = CreateFormatSchemaRequest{} }
func (m *CreateFormatSchemaRequest) String() string { return proto.CompactTextString(m) }
func (*CreateFormatSchemaRequest) ProtoMessage()    {}
func (*CreateFormatSchemaRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{3}
}

func (m *CreateFormatSchemaRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateFormatSchemaRequest.Unmarshal(m, b)
}
func (m *CreateFormatSchemaRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateFormatSchemaRequest.Marshal(b, m, deterministic)
}
func (m *CreateFormatSchemaRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateFormatSchemaRequest.Merge(m, src)
}
func (m *CreateFormatSchemaRequest) XXX_Size() int {
	return xxx_messageInfo_CreateFormatSchemaRequest.Size(m)
}
func (m *CreateFormatSchemaRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateFormatSchemaRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateFormatSchemaRequest proto.InternalMessageInfo

func (m *CreateFormatSchemaRequest) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *CreateFormatSchemaRequest) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

func (m *CreateFormatSchemaRequest) GetType() FormatSchemaType {
	if m != nil {
		return m.Type
	}
	return FormatSchemaType_FORMAT_SCHEMA_TYPE_UNSPECIFIED
}

func (m *CreateFormatSchemaRequest) GetUri() string {
	if m != nil {
		return m.Uri
	}
	return ""
}

type CreateFormatSchemaMetadata struct {
	ClusterId            string   `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string   `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateFormatSchemaMetadata) Reset()         { *m = CreateFormatSchemaMetadata{} }
func (m *CreateFormatSchemaMetadata) String() string { return proto.CompactTextString(m) }
func (*CreateFormatSchemaMetadata) ProtoMessage()    {}
func (*CreateFormatSchemaMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{4}
}

func (m *CreateFormatSchemaMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateFormatSchemaMetadata.Unmarshal(m, b)
}
func (m *CreateFormatSchemaMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateFormatSchemaMetadata.Marshal(b, m, deterministic)
}
func (m *CreateFormatSchemaMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateFormatSchemaMetadata.Merge(m, src)
}
func (m *CreateFormatSchemaMetadata) XXX_Size() int {
	return xxx_messageInfo_CreateFormatSchemaMetadata.Size(m)
}
func (m *CreateFormatSchemaMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateFormatSchemaMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_CreateFormatSchemaMetadata proto.InternalMessageInfo

func (m *CreateFormatSchemaMetadata) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *CreateFormatSchemaMetadata) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

type UpdateFormatSchemaRequest struct {
	ClusterId            string                `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string                `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	UpdateMask           *field_mask.FieldMask `protobuf:"bytes,3,opt,name=update_mask,json=updateMask,proto3" json:"update_mask,omitempty"`
	Uri                  string                `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *UpdateFormatSchemaRequest) Reset()         { *m = UpdateFormatSchemaRequest{} }
func (m *UpdateFormatSchemaRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateFormatSchemaRequest) ProtoMessage()    {}
func (*UpdateFormatSchemaRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{5}
}

func (m *UpdateFormatSchemaRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateFormatSchemaRequest.Unmarshal(m, b)
}
func (m *UpdateFormatSchemaRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateFormatSchemaRequest.Marshal(b, m, deterministic)
}
func (m *UpdateFormatSchemaRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateFormatSchemaRequest.Merge(m, src)
}
func (m *UpdateFormatSchemaRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateFormatSchemaRequest.Size(m)
}
func (m *UpdateFormatSchemaRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateFormatSchemaRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateFormatSchemaRequest proto.InternalMessageInfo

func (m *UpdateFormatSchemaRequest) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *UpdateFormatSchemaRequest) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

func (m *UpdateFormatSchemaRequest) GetUpdateMask() *field_mask.FieldMask {
	if m != nil {
		return m.UpdateMask
	}
	return nil
}

func (m *UpdateFormatSchemaRequest) GetUri() string {
	if m != nil {
		return m.Uri
	}
	return ""
}

type UpdateFormatSchemaMetadata struct {
	ClusterId            string   `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string   `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateFormatSchemaMetadata) Reset()         { *m = UpdateFormatSchemaMetadata{} }
func (m *UpdateFormatSchemaMetadata) String() string { return proto.CompactTextString(m) }
func (*UpdateFormatSchemaMetadata) ProtoMessage()    {}
func (*UpdateFormatSchemaMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{6}
}

func (m *UpdateFormatSchemaMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateFormatSchemaMetadata.Unmarshal(m, b)
}
func (m *UpdateFormatSchemaMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateFormatSchemaMetadata.Marshal(b, m, deterministic)
}
func (m *UpdateFormatSchemaMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateFormatSchemaMetadata.Merge(m, src)
}
func (m *UpdateFormatSchemaMetadata) XXX_Size() int {
	return xxx_messageInfo_UpdateFormatSchemaMetadata.Size(m)
}
func (m *UpdateFormatSchemaMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateFormatSchemaMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateFormatSchemaMetadata proto.InternalMessageInfo

func (m *UpdateFormatSchemaMetadata) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *UpdateFormatSchemaMetadata) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

type DeleteFormatSchemaRequest struct {
	ClusterId            string   `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string   `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteFormatSchemaRequest) Reset()         { *m = DeleteFormatSchemaRequest{} }
func (m *DeleteFormatSchemaRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteFormatSchemaRequest) ProtoMessage()    {}
func (*DeleteFormatSchemaRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{7}
}

func (m *DeleteFormatSchemaRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteFormatSchemaRequest.Unmarshal(m, b)
}
func (m *DeleteFormatSchemaRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteFormatSchemaRequest.Marshal(b, m, deterministic)
}
func (m *DeleteFormatSchemaRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteFormatSchemaRequest.Merge(m, src)
}
func (m *DeleteFormatSchemaRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteFormatSchemaRequest.Size(m)
}
func (m *DeleteFormatSchemaRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteFormatSchemaRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteFormatSchemaRequest proto.InternalMessageInfo

func (m *DeleteFormatSchemaRequest) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *DeleteFormatSchemaRequest) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

type DeleteFormatSchemaMetadata struct {
	ClusterId            string   `protobuf:"bytes,1,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	FormatSchemaName     string   `protobuf:"bytes,2,opt,name=format_schema_name,json=formatSchemaName,proto3" json:"format_schema_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteFormatSchemaMetadata) Reset()         { *m = DeleteFormatSchemaMetadata{} }
func (m *DeleteFormatSchemaMetadata) String() string { return proto.CompactTextString(m) }
func (*DeleteFormatSchemaMetadata) ProtoMessage()    {}
func (*DeleteFormatSchemaMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_7a66ff94ae18f1fd, []int{8}
}

func (m *DeleteFormatSchemaMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteFormatSchemaMetadata.Unmarshal(m, b)
}
func (m *DeleteFormatSchemaMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteFormatSchemaMetadata.Marshal(b, m, deterministic)
}
func (m *DeleteFormatSchemaMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteFormatSchemaMetadata.Merge(m, src)
}
func (m *DeleteFormatSchemaMetadata) XXX_Size() int {
	return xxx_messageInfo_DeleteFormatSchemaMetadata.Size(m)
}
func (m *DeleteFormatSchemaMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteFormatSchemaMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteFormatSchemaMetadata proto.InternalMessageInfo

func (m *DeleteFormatSchemaMetadata) GetClusterId() string {
	if m != nil {
		return m.ClusterId
	}
	return ""
}

func (m *DeleteFormatSchemaMetadata) GetFormatSchemaName() string {
	if m != nil {
		return m.FormatSchemaName
	}
	return ""
}

func init() {
	proto.RegisterType((*GetFormatSchemaRequest)(nil), "yandex.cloud.mdb.clickhouse.v1.GetFormatSchemaRequest")
	proto.RegisterType((*ListFormatSchemasRequest)(nil), "yandex.cloud.mdb.clickhouse.v1.ListFormatSchemasRequest")
	proto.RegisterType((*ListFormatSchemasResponse)(nil), "yandex.cloud.mdb.clickhouse.v1.ListFormatSchemasResponse")
	proto.RegisterType((*CreateFormatSchemaRequest)(nil), "yandex.cloud.mdb.clickhouse.v1.CreateFormatSchemaRequest")
	proto.RegisterType((*CreateFormatSchemaMetadata)(nil), "yandex.cloud.mdb.clickhouse.v1.CreateFormatSchemaMetadata")
	proto.RegisterType((*UpdateFormatSchemaRequest)(nil), "yandex.cloud.mdb.clickhouse.v1.UpdateFormatSchemaRequest")
	proto.RegisterType((*UpdateFormatSchemaMetadata)(nil), "yandex.cloud.mdb.clickhouse.v1.UpdateFormatSchemaMetadata")
	proto.RegisterType((*DeleteFormatSchemaRequest)(nil), "yandex.cloud.mdb.clickhouse.v1.DeleteFormatSchemaRequest")
	proto.RegisterType((*DeleteFormatSchemaMetadata)(nil), "yandex.cloud.mdb.clickhouse.v1.DeleteFormatSchemaMetadata")
}

func init() {
	proto.RegisterFile("yandex/cloud/mdb/clickhouse/v1/format_schema_service.proto", fileDescriptor_7a66ff94ae18f1fd)
}

var fileDescriptor_7a66ff94ae18f1fd = []byte{
	// 831 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x96, 0xcd, 0x4f, 0x1b, 0x47,
	0x18, 0xc6, 0x35, 0xd8, 0xb5, 0xf0, 0xf0, 0x51, 0x34, 0x55, 0x91, 0x6d, 0x15, 0x44, 0xf7, 0x40,
	0x91, 0x8b, 0x77, 0xbd, 0x46, 0x45, 0xe5, 0x4b, 0x55, 0xdd, 0x16, 0xd4, 0x0a, 0x4a, 0xbb, 0xa6,
	0xaa, 0x44, 0x55, 0x59, 0x63, 0xef, 0xd8, 0xac, 0xec, 0xfd, 0xa8, 0x67, 0xd6, 0xc2, 0x20, 0x2e,
	0x3d, 0x45, 0x1c, 0x72, 0x41, 0x8a, 0xa2, 0xfc, 0x19, 0xf9, 0x07, 0x72, 0x89, 0x04, 0xe7, 0xe4,
	0x98, 0x6b, 0x22, 0x45, 0xb9, 0x25, 0x97, 0x48, 0x9c, 0xa2, 0x9d, 0x31, 0x78, 0x37, 0x8b, 0x6d,
	0x20, 0x20, 0x71, 0xdb, 0xdd, 0xf7, 0x7d, 0x67, 0x9f, 0xdf, 0xb3, 0x3b, 0xcf, 0x2e, 0x5c, 0x6c,
	0x61, 0x4b, 0x27, 0xbb, 0x4a, 0xb9, 0x6e, 0xbb, 0xba, 0x62, 0xea, 0x25, 0xa5, 0x5c, 0x37, 0xca,
	0xb5, 0x1d, 0xdb, 0xa5, 0x44, 0x69, 0xaa, 0x4a, 0xc5, 0x6e, 0x98, 0x98, 0x15, 0x69, 0x79, 0x87,
	0x98, 0xb8, 0x48, 0x49, 0xa3, 0x69, 0x94, 0x89, 0xec, 0x34, 0x6c, 0x66, 0xa3, 0x49, 0x31, 0x2b,
	0xf3, 0x59, 0xd9, 0xd4, 0x4b, 0x72, 0x67, 0x56, 0x6e, 0xaa, 0xa9, 0xaf, 0xaa, 0xb6, 0x5d, 0xad,
	0x13, 0x05, 0x3b, 0x86, 0x82, 0x2d, 0xcb, 0x66, 0x98, 0x19, 0xb6, 0x45, 0xc5, 0x74, 0x6a, 0xaa,
	0x5d, 0xe5, 0x67, 0x25, 0xb7, 0xa2, 0x54, 0x0c, 0x52, 0xd7, 0x8b, 0x26, 0xa6, 0xb5, 0xb3, 0x8e,
	0x80, 0x36, 0x6f, 0x15, 0xdb, 0x21, 0x0d, 0xbe, 0x48, 0xbb, 0x63, 0x3a, 0xd0, 0x71, 0x5e, 0x0d,
	0xf5, 0x4d, 0x04, 0xfa, 0x9a, 0xb8, 0x6e, 0xe8, 0xfe, 0x72, 0xee, 0x2a, 0x26, 0x88, 0x19, 0xe9,
	0x08, 0xc0, 0xf1, 0x35, 0xc2, 0x56, 0x79, 0xa9, 0xc0, 0x2b, 0x1a, 0xf9, 0xcf, 0x25, 0x94, 0xa1,
	0x6f, 0x21, 0x2c, 0xd7, 0x5d, 0xca, 0x48, 0xa3, 0x68, 0xe8, 0x09, 0x30, 0x05, 0x66, 0xe2, 0xf9,
	0xe1, 0xd7, 0xc7, 0x2a, 0x38, 0x3c, 0x51, 0xa3, 0xcb, 0x2b, 0xdf, 0x65, 0xb5, 0x78, 0xbb, 0xfe,
	0xab, 0x8e, 0xd6, 0x21, 0x0a, 0x7a, 0x6c, 0x61, 0x93, 0x24, 0x06, 0xf8, 0xd0, 0xa4, 0x37, 0xf4,
	0xee, 0x58, 0x1d, 0xfd, 0x07, 0x67, 0xf6, 0x7e, 0xcc, 0x6c, 0x67, 0x33, 0x0b, 0xc5, 0xcc, 0xbf,
	0x69, 0xb1, 0xcc, 0xfc, 0x9c, 0x36, 0x56, 0xf1, 0xdd, 0xfd, 0x77, 0x6c, 0x12, 0xe9, 0x11, 0x80,
	0x89, 0x75, 0x83, 0x06, 0x64, 0xd1, 0x6b, 0xe9, 0xfa, 0x06, 0xc6, 0x1d, 0x5c, 0x25, 0x45, 0x6a,
	0xec, 0x09, 0x39, 0x91, 0x3c, 0x3c, 0x3d, 0x56, 0x63, 0xcb, 0x2b, 0x6a, 0x36, 0x9b, 0xd5, 0x06,
	0xbd, 0x62, 0xc1, 0xd8, 0x23, 0x68, 0x06, 0x42, 0xde, 0xc8, 0xec, 0x1a, 0xb1, 0x12, 0x11, 0xbe,
	0x6a, 0xfc, 0xf0, 0x44, 0xfd, 0x8c, 0x77, 0x6a, 0x7c, 0x95, 0x2d, 0xaf, 0x26, 0x3d, 0x04, 0x30,
	0x79, 0x81, 0x38, 0xea, 0xd8, 0x16, 0x25, 0xa8, 0x00, 0x47, 0x03, 0x46, 0xd0, 0x04, 0x98, 0x8a,
	0xcc, 0x0c, 0xe5, 0x66, 0xe5, 0xde, 0xaf, 0x99, 0x1c, 0x78, 0x04, 0x23, 0x7e, 0x4b, 0x28, 0x9a,
	0x86, 0x9f, 0x5b, 0x64, 0x97, 0x15, 0x7d, 0x0a, 0xb9, 0xb5, 0xda, 0x88, 0x77, 0xf9, 0x8f, 0x73,
	0x69, 0xa7, 0x00, 0x26, 0x7f, 0x6a, 0x10, 0xcc, 0xc8, 0xdd, 0x7a, 0xa0, 0xe8, 0x37, 0x18, 0x65,
	0x2d, 0x87, 0x70, 0x5f, 0x47, 0x73, 0xd9, 0xab, 0x78, 0xb1, 0xd5, 0x72, 0x48, 0x3e, 0xea, 0xdd,
	0x51, 0xe3, 0x6b, 0xa0, 0x71, 0x18, 0x71, 0x1b, 0x46, 0x22, 0xca, 0xa5, 0x88, 0x82, 0x77, 0x41,
	0x32, 0x60, 0x2a, 0xcc, 0xbe, 0x41, 0x18, 0xd6, 0x31, 0xc3, 0x68, 0x22, 0x0c, 0xef, 0xc7, 0x9d,
	0xed, 0x8e, 0x7b, 0xc1, 0xfb, 0xf9, 0x0a, 0xc0, 0xe4, 0x5f, 0x8e, 0x7e, 0xf7, 0x7c, 0x5e, 0x82,
	0x43, 0x2e, 0xd7, 0xc5, 0x03, 0x88, 0xdb, 0x3d, 0x94, 0x4b, 0xc9, 0x22, 0xa3, 0xe4, 0xb3, 0x8c,
	0x92, 0x57, 0xbd, 0x8c, 0xda, 0xc0, 0xb4, 0xa6, 0x41, 0xd1, 0xee, 0x1d, 0xa3, 0x31, 0x9f, 0xb1,
	0xe7, 0x96, 0x86, 0x31, 0x6f, 0xc7, 0xd2, 0x07, 0x00, 0x26, 0x7f, 0x26, 0x75, 0x72, 0xd7, 0x2c,
	0xf5, 0x3c, 0x08, 0xeb, 0xba, 0x15, 0x0f, 0x72, 0x6f, 0x06, 0xe1, 0x17, 0xfe, 0xbb, 0x14, 0xc4,
	0x77, 0x0a, 0x3d, 0x05, 0x30, 0xb2, 0x46, 0x18, 0x9a, 0xef, 0xb7, 0x6f, 0x2e, 0x4e, 0xf2, 0xd4,
	0x95, 0xb2, 0x47, 0xfa, 0xfb, 0xff, 0xe7, 0x2f, 0x8f, 0x06, 0xfe, 0x44, 0x9b, 0x8a, 0x89, 0x2d,
	0x5c, 0x25, 0x7a, 0x26, 0xf8, 0x19, 0x69, 0xd3, 0x51, 0x65, 0xbf, 0x43, 0x7e, 0xa0, 0x04, 0x02,
	0x4b, 0xd9, 0x0f, 0x53, 0x1f, 0xa0, 0x27, 0x00, 0x46, 0xbd, 0xe0, 0x44, 0xdf, 0xf7, 0xd3, 0xd3,
	0x2d, 0xfb, 0x53, 0x0b, 0xd7, 0x98, 0x14, 0xc1, 0x2c, 0xe5, 0x39, 0xd6, 0x32, 0x5a, 0xbc, 0x3e,
	0x16, 0x7a, 0x01, 0x60, 0x4c, 0x64, 0x0c, 0xea, 0xab, 0xa4, 0x6b, 0x0e, 0xa7, 0xbe, 0x0e, 0x8e,
	0x76, 0xbe, 0xf2, 0x9b, 0x67, 0x47, 0x12, 0x7d, 0xfc, 0x2c, 0x9d, 0xee, 0x99, 0x67, 0xc3, 0xfe,
	0xab, 0x1c, 0xed, 0x07, 0xe9, 0x13, 0xd0, 0x16, 0x41, 0x1a, 0xbd, 0x05, 0x30, 0x26, 0xb6, 0x7b,
	0x7f, 0xba, 0xae, 0xe9, 0x77, 0x19, 0xba, 0x7b, 0x40, 0xe0, 0xf5, 0xc8, 0x96, 0x30, 0xde, 0x56,
	0xee, 0xa6, 0x5f, 0x48, 0x8f, 0xf9, 0x3d, 0x80, 0x31, 0xb1, 0xbd, 0xfb, 0x33, 0x77, 0x8d, 0xa7,
	0xcb, 0x30, 0xdf, 0xf7, 0x98, 0xe7, 0x7a, 0x66, 0xc9, 0x97, 0x1f, 0xc7, 0xf4, 0x2f, 0xa6, 0xc3,
	0x5a, 0x62, 0x37, 0xa6, 0x6f, 0x1a, 0x3e, 0x4f, 0xa1, 0x14, 0x10, 0x8d, 0x1d, 0x23, 0xcc, 0xbc,
	0xbd, 0x51, 0x35, 0xd8, 0x8e, 0x5b, 0x92, 0xcb, 0xb6, 0xa9, 0x88, 0xf6, 0x8c, 0xf8, 0xbd, 0xac,
	0xda, 0x99, 0x2a, 0xb1, 0xb8, 0x54, 0xa5, 0xf7, 0x7f, 0xe7, 0x52, 0xe7, 0xac, 0x14, 0xe3, 0x03,
	0x73, 0x1f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x98, 0x05, 0xc8, 0x5b, 0xb0, 0x0b, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// FormatSchemaServiceClient is the client API for FormatSchemaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FormatSchemaServiceClient interface {
	Get(ctx context.Context, in *GetFormatSchemaRequest, opts ...grpc.CallOption) (*FormatSchema, error)
	List(ctx context.Context, in *ListFormatSchemasRequest, opts ...grpc.CallOption) (*ListFormatSchemasResponse, error)
	Create(ctx context.Context, in *CreateFormatSchemaRequest, opts ...grpc.CallOption) (*operation.Operation, error)
	Update(ctx context.Context, in *UpdateFormatSchemaRequest, opts ...grpc.CallOption) (*operation.Operation, error)
	Delete(ctx context.Context, in *DeleteFormatSchemaRequest, opts ...grpc.CallOption) (*operation.Operation, error)
}

type formatSchemaServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFormatSchemaServiceClient(cc grpc.ClientConnInterface) FormatSchemaServiceClient {
	return &formatSchemaServiceClient{cc}
}

func (c *formatSchemaServiceClient) Get(ctx context.Context, in *GetFormatSchemaRequest, opts ...grpc.CallOption) (*FormatSchema, error) {
	out := new(FormatSchema)
	err := c.cc.Invoke(ctx, "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *formatSchemaServiceClient) List(ctx context.Context, in *ListFormatSchemasRequest, opts ...grpc.CallOption) (*ListFormatSchemasResponse, error) {
	out := new(ListFormatSchemasResponse)
	err := c.cc.Invoke(ctx, "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *formatSchemaServiceClient) Create(ctx context.Context, in *CreateFormatSchemaRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	out := new(operation.Operation)
	err := c.cc.Invoke(ctx, "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *formatSchemaServiceClient) Update(ctx context.Context, in *UpdateFormatSchemaRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	out := new(operation.Operation)
	err := c.cc.Invoke(ctx, "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *formatSchemaServiceClient) Delete(ctx context.Context, in *DeleteFormatSchemaRequest, opts ...grpc.CallOption) (*operation.Operation, error) {
	out := new(operation.Operation)
	err := c.cc.Invoke(ctx, "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FormatSchemaServiceServer is the server API for FormatSchemaService service.
type FormatSchemaServiceServer interface {
	Get(context.Context, *GetFormatSchemaRequest) (*FormatSchema, error)
	List(context.Context, *ListFormatSchemasRequest) (*ListFormatSchemasResponse, error)
	Create(context.Context, *CreateFormatSchemaRequest) (*operation.Operation, error)
	Update(context.Context, *UpdateFormatSchemaRequest) (*operation.Operation, error)
	Delete(context.Context, *DeleteFormatSchemaRequest) (*operation.Operation, error)
}

// UnimplementedFormatSchemaServiceServer can be embedded to have forward compatible implementations.
type UnimplementedFormatSchemaServiceServer struct {
}

func (*UnimplementedFormatSchemaServiceServer) Get(ctx context.Context, req *GetFormatSchemaRequest) (*FormatSchema, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedFormatSchemaServiceServer) List(ctx context.Context, req *ListFormatSchemasRequest) (*ListFormatSchemasResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (*UnimplementedFormatSchemaServiceServer) Create(ctx context.Context, req *CreateFormatSchemaRequest) (*operation.Operation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (*UnimplementedFormatSchemaServiceServer) Update(ctx context.Context, req *UpdateFormatSchemaRequest) (*operation.Operation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (*UnimplementedFormatSchemaServiceServer) Delete(ctx context.Context, req *DeleteFormatSchemaRequest) (*operation.Operation, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func RegisterFormatSchemaServiceServer(s *grpc.Server, srv FormatSchemaServiceServer) {
	s.RegisterService(&_FormatSchemaService_serviceDesc, srv)
}

func _FormatSchemaService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetFormatSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FormatSchemaServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FormatSchemaServiceServer).Get(ctx, req.(*GetFormatSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FormatSchemaService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListFormatSchemasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FormatSchemaServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FormatSchemaServiceServer).List(ctx, req.(*ListFormatSchemasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FormatSchemaService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateFormatSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FormatSchemaServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FormatSchemaServiceServer).Create(ctx, req.(*CreateFormatSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FormatSchemaService_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateFormatSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FormatSchemaServiceServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FormatSchemaServiceServer).Update(ctx, req.(*UpdateFormatSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FormatSchemaService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteFormatSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FormatSchemaServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yandex.cloud.mdb.clickhouse.v1.FormatSchemaService/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FormatSchemaServiceServer).Delete(ctx, req.(*DeleteFormatSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _FormatSchemaService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "yandex.cloud.mdb.clickhouse.v1.FormatSchemaService",
	HandlerType: (*FormatSchemaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _FormatSchemaService_Get_Handler,
		},
		{
			MethodName: "List",
			Handler:    _FormatSchemaService_List_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _FormatSchemaService_Create_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _FormatSchemaService_Update_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _FormatSchemaService_Delete_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "yandex/cloud/mdb/clickhouse/v1/format_schema_service.proto",
}
