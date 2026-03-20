package services

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mcache-team/mcache/pkg/proto"
	"github.com/mcache-team/mcache/pkg/services/grpcserver"
	"github.com/mcache-team/mcache/pkg/services/rest"
	"github.com/mcache-team/mcache/pkg/storage"
	"github.com/mcache-team/mcache/pkg/storage/memory"
	"github.com/mcache-team/mcache/pkg/storage/persist"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// compile-time check
var _ persist.Snapshotter = &memory.Memory{}

func Start() {
	gin.SetMode(gin.DebugMode)
	initPersist()
	go startGRPC()
	startHTTP()
}

func initPersist() {
	dir := os.Getenv("PERSIST_DIR")
	if dir == "" {
		return
	}
	interval, err := time.ParseDuration(os.Getenv("PERSIST_INTERVAL"))
	if err != nil || interval <= 0 {
		interval = 5 * time.Minute
	}
	snap, ok := storage.StorageClient.Store().(persist.Snapshotter)
	if !ok {
		logrus.Warn("persist: storage does not support snapshots, skipping")
		return
	}
	mgr := persist.New(snap, dir, interval)
	if err := mgr.Load(); err != nil {
		logrus.Errorf("persist: load failed: %v", err)
	}
	go mgr.Start()
	logrus.Infof("persist: enabled, dir=%s interval=%s", dir, interval)
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
