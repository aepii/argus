package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aepii/argus"
	"github.com/aepii/argus/internal/embed"
	"github.com/aepii/argus/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestConcurrentLoad(t *testing.T) {
	const numGoroutines = 20

	config, err := loadConfig("../../.env")
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	address := config.address
	port := config.port
	target := address + ":" + port

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient failed: %v", err)
	}
	defer conn.Close()

	client := pb.NewShardServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	embedClient, err := embed.NewClient(config.endpoint, config.apiKey, config.apiVersion, config.model, config.dim)
	if err != nil {
		t.Fatalf("embed.NewClient failed: %v", err)
	}

	items := make([][]argus.UpsertItem, numGoroutines)
	rawTexts := []string{"I've waited years.",
		"How... How many years?",
		"By now it must be...",
		"It's been twenty-three years, four months, eight days.",
		"Hello world!",
		"Now I am become Death, the destroyer of worlds.",
		"Fist my bump!",
		"Fill your belly. Day and night make merry. Let days be full of joy... For these alone are the concerns of man.",
		"I keep on moving forward until I exterminate my enemies."}
	for workerID := range numGoroutines {
		for i, rawText := range rawTexts {
			id := workerID*len(rawTexts) + i
			rawText := fmt.Sprintf("%s (%d)", rawText, id)
			items[workerID] = append(items[workerID], argus.UpsertItem{ID: int64(id), RawText: rawText, Embedding: nil})
		}
	}

	var wg sync.WaitGroup

	successfulGoroutines := make([]bool, numGoroutines)
	for workerID := 0; workerID < numGoroutines; workerID++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			workerItems := items[workerID]
			batchToEmbed := make([]string, len(workerItems))
			for i, workerItem := range workerItems {
				batchToEmbed[i] = workerItem.RawText
			}

			embeds, err := embedClient.EmbedBatch(ctx, batchToEmbed)
			if err != nil {
				t.Errorf("EmbedBatch failed: %v", err)
				return
			}

			pbItems := make([]*pb.UpsertItem, len(workerItems))
			for i := range workerItems {
				workerItem := workerItems[i]
				workerItems[i].Embedding = embeds[i]

				rawText := workerItem.RawText
				id := workerItem.ID

				pbItems[i] = &pb.UpsertItem{Id: int64(id), RawText: rawText, Embedding: embeds[i]}
			}

			_, err = client.UpsertBatch(ctx, &pb.UpsertBatchRequest{Items: pbItems})
			if err != nil {
				t.Errorf("UpsertBatch failed: %v", err)
				return
			}
			successfulGoroutines[workerID] = true
		}(workerID)
	}

	wg.Wait()

	for workerID, workerItems := range items {
		if !successfulGoroutines[workerID] {
			continue
		}
		for _, item := range workerItems {
			t.Run("search/"+item.RawText, func(t *testing.T) {
				search, err := client.Search(ctx, &pb.SearchRequest{Embedding: item.Embedding, TopK: 1})
				if err != nil {
					t.Fatalf("Search failed: %v", err)
				}

				if len(search.Results) != 1 || search.Results[0].RawText != item.RawText {
					t.Errorf("Search() = %v, want 1 result with RawText %q", search, item.RawText)
				}
			})
		}
	}
}
