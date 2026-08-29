package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/aepii/argus"
	"github.com/aepii/argus/internal/config"
	"github.com/aepii/argus/pb"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadServer("../../.env")
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	port := cfg.Port
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	path := cfg.DbPath
	dim := cfg.EmbedDim

	store, err := argus.NewVectorStore(path, dim)
	if err != nil {
		slog.Error("failed to create vector store", "error", err)
		os.Exit(1)
	}
	shardServer, err := argus.NewShardServer(store)
	if err != nil {
		slog.Error("failed to create shard server", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterShardServiceServer(grpcServer, shardServer)
	slog.Info("server is listening", "port", port)
	err = grpcServer.Serve(lis)
	if err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
