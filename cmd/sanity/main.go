package main

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"

	"github.com/aepii/argus"
)

func randomEmbedding(dim int) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = rand.Float32()*2 - 1
	}
	return v
}

const embeddingDim = 384

func main() {
	s, err := argus.NewVectorStore(".test.db", embeddingDim)
	if err != nil {
		slog.Error("failed to create vector store", "error", err)
		os.Exit(1)
	}

	s.Upsert(1, "hello!", randomEmbedding(embeddingDim))
	s.Upsert(10, "world!", randomEmbedding(embeddingDim))
	s.Upsert(5, "hi!", randomEmbedding(embeddingDim))
	s.Upsert(1, "bye!", randomEmbedding(embeddingDim))

	res, err := s.Search(randomEmbedding(embeddingDim), 3)
	if err != nil {
		slog.Error("failed to search vector store", "error", err)
		os.Exit(1)
	}
	fmt.Println(res)

	s.Count()
}
