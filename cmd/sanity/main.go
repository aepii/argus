package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aepii/argus"
	"github.com/aepii/argus/internal/config"
	"github.com/aepii/argus/internal/embed"
)

func main() {
	cfg, err := config.LoadSanity("../../.env")
	if err != nil {
		slog.Error("failed to load environment variables", "error", err)
		os.Exit(1)
	}

	s, err := argus.NewVectorStore(cfg.DbPath, cfg.EmbedDim)
	if err != nil {
		slog.Error("failed to create vector store", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	embedClient, err := embed.NewClient(cfg.Endpoint, cfg.APIKey, cfg.APIVersion, cfg.Model, cfg.EmbedDim)
	if err != nil {
		slog.Error("failed to create embedding client", "error", err)
		os.Exit(1)
	}

	var strings = []string{"Hello World!", "Bye World!", "The sky is blue!", "I love dogs!", "I love cats!", "I hate animals!", "I feel horrible typing that previous sentence.", "The sky is really green!"}

	embs, err := embedClient.EmbedBatch(ctx, strings)
	if err != nil {
		slog.Error("failed to create embedding", "error", err)
		os.Exit(1)
	}

	for id, emb := range embs {
		id := int64(id)
		rawText := strings[id]
		err = s.Upsert(argus.UpsertItem{ID: id, RawText: rawText, Embedding: emb})
		if err != nil {
			slog.Error("failed to upsert embedding", "error", err, "id", id, "rawText", rawText, "emb", emb)
			os.Exit(1)
		}
	}

	c, err := s.Count()
	if err != nil {
		slog.Error("failed to count", "error", err)
		os.Exit(1)
	}
	fmt.Printf("count: %d\n", c)

	const searchString = "The sky is green!"
	emb, err := embedClient.Embed(ctx, searchString)
	if err != nil {
		slog.Error("failed to create embedding", "error", err)
		os.Exit(1)
	}

	res, err := s.Search(emb, 3)
	if err != nil {
		slog.Error("failed to search embedding", "error", err, "emb", emb)
		os.Exit(1)
	}
	fmt.Println(res)
}
