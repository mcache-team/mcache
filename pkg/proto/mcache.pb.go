// Code generated manually (proto: mcache.proto)
// source: pkg/proto/mcache.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

type InsertRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix     string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Data       []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	TtlSeconds int64  `protobuf:"varint,3,opt,name=ttl_seconds,json=ttlSeconds,proto3" json:"ttl_seconds,omitempty"`
}

func (x *InsertRequest) Reset()         { *x = InsertRequest{} }
func (x *InsertRequest) String() string  { return x.Prefix }
func (x *InsertRequest) ProtoMessage()   {}
func (x *InsertRequest) GetPrefix() string { return x.Prefix }
func (x *InsertRequest) GetData() []byte   { return x.Data }
func (x *InsertRequest) GetTtlSeconds() int64 { return x.TtlSeconds }
func (x *InsertRequest) ProtoReflect() protoreflect.Message {
	return nil
}

type InsertResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *InsertResponse) Reset()         { *x = InsertResponse{} }
func (x *InsertResponse) String() string  { return "" }
func (x *InsertResponse) ProtoMessage()   {}
func (x *InsertResponse) GetSuccess() bool { return x.Success }
func (x *InsertResponse) ProtoReflect() protoreflect.Message {
	return nil
}

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *GetRequest) Reset()           { *x = GetRequest{} }
func (x *GetRequest) String() string    { return x.Prefix }
func (x *GetRequest) ProtoMessage()     {}
func (x *GetRequest) GetPrefix() string { return x.Prefix }
func (x *GetRequest) ProtoReflect() protoreflect.Message {
	return nil
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix     string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Data       []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	ExpireTime int64  `protobuf:"varint,3,opt,name=expire_time,json=expireTime,proto3" json:"expire_time,omitempty"`
	CreatedAt  int64  `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt  int64  `protobuf:"varint,5,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (x *GetResponse) Reset()              { *x = GetResponse{} }
func (x *GetResponse) String() string       { return x.Prefix }
func (x *GetResponse) ProtoMessage()        {}
func (x *GetResponse) GetPrefix() string    { return x.Prefix }
func (x *GetResponse) GetData() []byte      { return x.Data }
func (x *GetResponse) GetExpireTime() int64 { return x.ExpireTime }
func (x *GetResponse) GetCreatedAt() int64  { return x.CreatedAt }
func (x *GetResponse) GetUpdatedAt() int64  { return x.UpdatedAt }
func (x *GetResponse) ProtoReflect() protoreflect.Message {
	return nil
}

type UpdateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix     string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Data       []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	TtlSeconds int64  `protobuf:"varint,3,opt,name=ttl_seconds,json=ttlSeconds,proto3" json:"ttl_seconds,omitempty"`
}

func (x *UpdateRequest) Reset()           { *x = UpdateRequest{} }
func (x *UpdateRequest) String() string    { return x.Prefix }
func (x *UpdateRequest) ProtoMessage()     {}
func (x *UpdateRequest) GetPrefix() string { return x.Prefix }
func (x *UpdateRequest) GetData() []byte   { return x.Data }
func (x *UpdateRequest) GetTtlSeconds() int64 { return x.TtlSeconds }
func (x *UpdateRequest) ProtoReflect() protoreflect.Message {
	return nil
}

type UpdateResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *UpdateResponse) Reset()         { *x = UpdateResponse{} }
func (x *UpdateResponse) String() string  { return "" }
func (x *UpdateResponse) ProtoMessage()   {}
func (x *UpdateResponse) GetSuccess() bool { return x.Success }
func (x *UpdateResponse) ProtoReflect() protoreflect.Message {
	return nil
}

type DeleteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *DeleteRequest) Reset()           { *x = DeleteRequest{} }
func (x *DeleteRequest) String() string    { return x.Prefix }
func (x *DeleteRequest) ProtoMessage()     {}
func (x *DeleteRequest) GetPrefix() string { return x.Prefix }
func (x *DeleteRequest) ProtoReflect() protoreflect.Message {
	return nil
}

type DeleteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *DeleteResponse) Reset()         { *x = DeleteResponse{} }
func (x *DeleteResponse) String() string  { return "" }
func (x *DeleteResponse) ProtoMessage()   {}
func (x *DeleteResponse) GetSuccess() bool { return x.Success }
func (x *DeleteResponse) ProtoReflect() protoreflect.Message {
	return nil
}

type ListByPrefixRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Prefix string `protobuf:"bytes,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
}

func (x *ListByPrefixRequest) Reset()           { *x = ListByPrefixRequest{} }
func (x *ListByPrefixRequest) String() string    { return x.Prefix }
func (x *ListByPrefixRequest) ProtoMessage()     {}
func (x *ListByPrefixRequest) GetPrefix() string { return x.Prefix }
func (x *ListByPrefixRequest) ProtoReflect() protoreflect.Message {
	return nil
}

type ListByPrefixResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*GetResponse `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListByPrefixResponse) Reset()              { *x = ListByPrefixResponse{} }
func (x *ListByPrefixResponse) String() string       { return "" }
func (x *ListByPrefixResponse) ProtoMessage()        {}
func (x *ListByPrefixResponse) GetItems() []*GetResponse { return x.Items }
func (x *ListByPrefixResponse) ProtoReflect() protoreflect.Message {
	return nil
}

var _ = sync.Once{}
var _ = reflect.TypeOf
