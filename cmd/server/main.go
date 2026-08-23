package main

import (
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/aepii/argus"
	"github.com/aepii/argus/pb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	path := os.Getenv("DB_PATH")
	dim := os.Getenv("EMBEDDING_DIM")
	dim64, err := strconv.ParseUint(dim, 10, 16)
	if err != nil {
		slog.Error("failed to convert string", "error", err, "dim", dim)
		os.Exit(1)
	}
	dim16 := uint16(dim64)

	store, err := argus.NewVectorStore(path, dim16)
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
