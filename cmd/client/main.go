package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/aepii/argus/internal/embed"
	"github.com/aepii/argus/pb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	endpoint   string
	apiKey     string
	apiVersion string
	model      string
	dim        uint16
}

func loadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		return nil, err
	}

	dim64, err := strconv.ParseUint(os.Getenv("EMBEDDING_DIM"), 10, 16)
	if err != nil {
		return nil, err
	}
	dim16 := uint16(dim64)

	return &Config{
		endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
		apiKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
		apiVersion: os.Getenv("AZURE_OPENAI_API_VERSION"),
		model:      os.Getenv("AZURE_OPENAI_MODEL"),
		dim:        dim16,
	}, nil
}

func main() {
	config, err := loadConfig()
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	address := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")
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

	embedClient, err := embed.NewClient(config.endpoint, config.apiKey, config.apiVersion, config.model, config.dim)
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
