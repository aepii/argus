package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aepii/argus"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/azure"
)

const azureOpenAIAPIVersion = "2023-05-15"
const embeddingDim = 384

func main() {
	godotenv.Load()

	s, err := argus.NewVectorStore(".test.db", embeddingDim)
	if err != nil {
		slog.Error("failed to create vector store", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := openai.NewClient(
		azure.WithEndpoint(os.Getenv("AZURE_OPENAI_ENDPOINT"), azureOpenAIAPIVersion),
		azure.WithAPIKey(os.Getenv("AZURE_OPENAI_API_KEY")),
	)

	var strings = []string{"Hello World!", "Bye World!", "The sky is blue!", "I love dogs!", "I love cats!", "I hate animals!", "I feel horrible typing that previous sentence.", "The sky is really green!"}

	response, err := client.Embeddings.New(
		ctx,
		openai.EmbeddingNewParams{
			Model: "text-embedding-3-small",
			Input: openai.EmbeddingNewParamsInputUnion{
				OfArrayOfStrings: strings,
			},
			Dimensions: openai.Int(embeddingDim),
		},
	)
	if err != nil {
		slog.Error("failed to create embedding", "error", err)
		os.Exit(1)
	}

	for id, data := range response.Data {
		id := int64(id)
		rawText := strings[id]
		emb := data.Embedding
		emb32 := make([]float32, len(emb))
		for i, v := range emb {
			emb32[i] = float32(v)
		}
		err = s.Upsert(id, rawText, emb32)
		if err != nil {
			slog.Error("failed to upsert embedding", "error", err, "id", id, "rawText", rawText, "emb", emb32)
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
	response, err = client.Embeddings.New(
		ctx,
		openai.EmbeddingNewParams{
			Model: "text-embedding-3-small",
			Input: openai.EmbeddingNewParamsInputUnion{
				OfString: openai.String(searchString),
			},
			Dimensions: openai.Int(embeddingDim),
		},
	)
	if err != nil {
		slog.Error("failed to create embedding", "error", err)
		os.Exit(1)
	}

	emb := response.Data[0].Embedding
	emb32 := make([]float32, len(emb))
	for i, v := range emb {
		emb32[i] = float32(v)
	}

	res, err := s.Search(emb32, 3)
	if err != nil {
		slog.Error("failed to search embedding", "error", err, "emb", emb32)
		os.Exit(1)
	}
	fmt.Println(res)
}
