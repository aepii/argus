package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/aepii/argus"
)

func main() {
	s, err := argus.NewVectorStore(":memory:")
	if err != nil {
		slog.Error("failed to create vector store", "error", err)
		os.Exit(1)
	}
	fmt.Println(s)
}
