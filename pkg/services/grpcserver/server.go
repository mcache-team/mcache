package grpcserver

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/mcache-team/mcache/pkg/apis/v1/item"
	"github.com/mcache-team/mcache/pkg/handlers"
	"github.com/mcache-team/mcache/pkg/proto"
	"github.com/mcache-team/mcache/pkg/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CacheServer struct {
	proto.UnimplementedCacheServiceServer
}

func New() *CacheServer {
	return &CacheServer{}
}

func (s *CacheServer) Insert(_ context.Context, req *proto.InsertRequest) (*proto.InsertResponse, error) {
	if req.Prefix == "" {
		return nil, status.Error(codes.InvalidArgument, "prefix is empty")
	}
	var data interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		// treat as raw string if not valid JSON
		data = string(req.Data)
	}
	var opts []item.Option
	if req.TtlSeconds > 0 {
		opts = append(opts, item.WithTTL(time.Duration(req.TtlSeconds)*time.Second))
	}
	if err := handlers.PrefixHandler.InsertNode(req.Prefix, data, opts...); err != nil {
		if errors.Is(err, item.PrefixExisted) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.InsertResponse{Success: true}, nil
}

func (s *CacheServer) Get(_ context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	it, err := storage.StorageClient.GetOne(req.Prefix)
	if err != nil {
		if errors.Is(err, item.NoDataError) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	data, _ := json.Marshal(it.Data)
	resp := &proto.GetResponse{
		Prefix:    it.Prefix,
		Data:      data,
		CreatedAt: it.CreatedAt.Unix(),
		UpdatedAt: it.UpdatedAt.Unix(),
	}
	if !it.ExpireTime.IsZero() {
		resp.ExpireTime = it.ExpireTime.Unix()
	}
	return resp, nil
}

func (s *CacheServer) Update(_ context.Context, req *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	var data interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		data = string(req.Data)
	}
	var opts []item.Option
	if req.TtlSeconds > 0 {
		opts = append(opts, item.WithTTL(time.Duration(req.TtlSeconds)*time.Second))
	}
	if err := storage.StorageClient.Update(req.Prefix, data, opts...); err != nil {
		if errors.Is(err, item.NoDataError) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.UpdateResponse{Success: true}, nil
}

func (s *CacheServer) Delete(_ context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	if err := handlers.PrefixHandler.RemoveNode(req.Prefix); err != nil {
		if errors.Is(err, item.PrefixNotExisted) || errors.Is(err, item.NoDataError) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &proto.DeleteResponse{Success: true}, nil
}

func (s *CacheServer) ListByPrefix(_ context.Context, req *proto.ListByPrefixRequest) (*proto.ListByPrefixResponse, error) {
	items, err := handlers.PrefixHandler.ListNode(req.Prefix)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &proto.ListByPrefixResponse{Items: make([]*proto.GetResponse, 0, len(items))}
	for _, it := range items {
		data, _ := json.Marshal(it.Data)
		r := &proto.GetResponse{
			Prefix:    it.Prefix,
			Data:      data,
			CreatedAt: it.CreatedAt.Unix(),
			UpdatedAt: it.UpdatedAt.Unix(),
		}
		if !it.ExpireTime.IsZero() {
			r.ExpireTime = it.ExpireTime.Unix()
		}
		resp.Items = append(resp.Items, r)
	}
	return resp, nil
}
