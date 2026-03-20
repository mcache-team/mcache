// Code generated manually (proto: mcache.proto)
// source: pkg/proto/mcache.proto

package proto

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// CacheServiceClient is the client API for CacheService service.
type CacheServiceClient interface {
	Insert(ctx context.Context, in *InsertRequest, opts ...grpc.CallOption) (*InsertResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*UpdateResponse, error)
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	ListByPrefix(ctx context.Context, in *ListByPrefixRequest, opts ...grpc.CallOption) (*ListByPrefixResponse, error)
}

type cacheServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCacheServiceClient(cc grpc.ClientConnInterface) CacheServiceClient {
	return &cacheServiceClient{cc}
}

func (c *cacheServiceClient) Insert(ctx context.Context, in *InsertRequest, opts ...grpc.CallOption) (*InsertResponse, error) {
	out := new(InsertResponse)
	err := c.cc.Invoke(ctx, "/mcache.CacheService/Insert", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/mcache.CacheService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheServiceClient) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*UpdateResponse, error) {
	out := new(UpdateResponse)
	err := c.cc.Invoke(ctx, "/mcache.CacheService/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, "/mcache.CacheService/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cacheServiceClient) ListByPrefix(ctx context.Context, in *ListByPrefixRequest, opts ...grpc.CallOption) (*ListByPrefixResponse, error) {
	out := new(ListByPrefixResponse)
	err := c.cc.Invoke(ctx, "/mcache.CacheService/ListByPrefix", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CacheServiceServer is the server API for CacheService service.
type CacheServiceServer interface {
	Insert(context.Context, *InsertRequest) (*InsertResponse, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Update(context.Context, *UpdateRequest) (*UpdateResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	ListByPrefix(context.Context, *ListByPrefixRequest) (*ListByPrefixResponse, error)
	mustEmbedUnimplementedCacheServiceServer()
}

// UnimplementedCacheServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCacheServiceServer struct{}

func (UnimplementedCacheServiceServer) Insert(context.Context, *InsertRequest) (*InsertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Insert not implemented")
}
func (UnimplementedCacheServiceServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedCacheServiceServer) Update(context.Context, *UpdateRequest) (*UpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedCacheServiceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedCacheServiceServer) ListByPrefix(context.Context, *ListByPrefixRequest) (*ListByPrefixResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListByPrefix not implemented")
}
func (UnimplementedCacheServiceServer) mustEmbedUnimplementedCacheServiceServer() {}

// UnsafeCacheServiceServer may be embedded to opt out of forward compatibility for this service.
type UnsafeCacheServiceServer interface {
	mustEmbedUnimplementedCacheServiceServer()
}

func RegisterCacheServiceServer(s grpc.ServiceRegistrar, srv CacheServiceServer) {
	s.RegisterService(&CacheService_ServiceDesc, srv)
}

func _CacheService_Insert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServiceServer).Insert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/mcache.CacheService/Insert"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServiceServer).Insert(ctx, req.(*InsertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CacheService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/mcache.CacheService/Get"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CacheService_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServiceServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/mcache.CacheService/Update"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServiceServer).Update(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CacheService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/mcache.CacheService/Delete"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServiceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CacheService_ListByPrefix_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListByPrefixRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CacheServiceServer).ListByPrefix(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/mcache.CacheService/ListByPrefix"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CacheServiceServer).ListByPrefix(ctx, req.(*ListByPrefixRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var CacheService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mcache.CacheService",
	HandlerType: (*CacheServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Insert", Handler: _CacheService_Insert_Handler},
		{MethodName: "Get", Handler: _CacheService_Get_Handler},
		{MethodName: "Update", Handler: _CacheService_Update_Handler},
		{MethodName: "Delete", Handler: _CacheService_Delete_Handler},
		{MethodName: "ListByPrefix", Handler: _CacheService_ListByPrefix_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mcache.proto",
}
