package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aepii/argus/internal/config"
	"github.com/aepii/argus/internal/embed"
	"github.com/aepii/argus/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.LoadClient("../../.env")
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	address := cfg.Address
	port := cfg.Port
	target := address + ":" + port

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	slog.Info("successfully initialized client connection", "target", target)

	client := pb.NewShardServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	embedClient, err := embed.NewClient(cfg.Endpoint, cfg.APIKey, cfg.APIVersion, cfg.Model, cfg.EmbedDim)
	if err != nil {
		slog.Error("failed to create embedding client", "error", err)
		os.Exit(1)
	}

	var rawTexts = []string{"Hello World!", "Bye World!", "The sky is blue!", "I love dogs!", "I love cats!", "I hate animals!", "I feel horrible typing that previous sentence.", "The sky is really green!"}
	embs, err := embedClient.EmbedBatch(ctx, rawTexts)
	if err != nil {
		slog.Error("failed to create embedding", "error", err)
		os.Exit(1)
	}

	pbItems := make([]*pb.UpsertItem, len(embs))
	for i, emb := range embs {
		pbItems[i] = &pb.UpsertItem{Id: int64(i), RawText: rawTexts[i], Embedding: emb}
	}

	_, err = client.UpsertBatch(ctx, &pb.UpsertBatchRequest{Items: pbItems})
	if err != nil {
		slog.Error("failed to upsert batch", "error", err)
		os.Exit(1)
	}

	res, err := client.Count(ctx, &pb.CountRequest{})
	if err != nil {
		slog.Error("failed to count", "error", err)
		os.Exit(1)
	}
	fmt.Println(res.Count)

	rawText := "Hello Blue World!"
	emb, err := embedClient.Embed(ctx, rawText)
	if err != nil {
		slog.Error("failed to create embedding", "error", err)
		os.Exit(1)
	}

	search, err := client.Search(ctx, &pb.SearchRequest{Embedding: emb, TopK: 3})
	if err != nil {
		slog.Error("failed to search", "error", err)
		os.Exit(1)
	}

	fmt.Println(search.Results)
}
