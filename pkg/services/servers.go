package services

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/proto"
	"github.com/mcache-team/mcache/pkg/services/grpcserver"
	"github.com/mcache-team/mcache/pkg/services/rest"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Start() {
	gin.SetMode(gin.DebugMode)

	go startGRPC()

	startHTTP()
}

func startHTTP() {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{})
	})
	rest.Init(engine)

	addr := "0.0.0.0:8080"
	logrus.Printf("HTTP server listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		logrus.Fatalf("HTTP server failed: %v", err)
	}
}

func startGRPC() {
	addr := "0.0.0.0:9090"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Fatalf("gRPC listen failed: %v", err)
	}
	s := grpc.NewServer()
	proto.RegisterCacheServiceServer(s, grpcserver.New())
	logrus.Printf("gRPC server listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("gRPC server failed: %v", err)
	}
}
