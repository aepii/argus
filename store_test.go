package argus

import (
	"testing"
)

const memoryPath = ":memory:"
const embeddingDim = 3

func TestUpsert(t *testing.T) {

	s, err := NewVectorStore(memoryPath, embeddingDim)
	if err != nil {
		t.Fatalf("NewVectorStore failed: %v", err)
	}

	emb := []float32{1, 0, 0}
	rawText := "Hello world!"

	err = s.Upsert(UpsertItem{1, rawText, emb})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	searchEmb := []float32{0.9, 0.1, -0.1}
	res, err := s.Search(searchEmb, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(res) != 1 || res[0].RawText != rawText {
		t.Errorf("Search() = %v, want 1 result with RawText %q", res, rawText)
	}
}

func TestBatchUpsert(t *testing.T) {
	s, err := NewVectorStore(memoryPath, embeddingDim)
	if err != nil {
		t.Fatalf("NewVectorStore failed: %v", err)
	}

	var tests = []struct {
		id   int64
		text string
		emb  []float32
	}{
		{1, "Hello world!", []float32{1, 0, 0}},
		{2, "Bye world!", []float32{0.9, 0.1, -0.1}},
		{3, "Quack!", []float32{0.1, 0.1, -1}},
		{4, "Duck!", []float32{0.05, 0.15, -0.9}},
		{5, "The sky is blue!", []float32{0.45, 0.5, 0.6}},
	}

	for _, tc := range tests {
		t.Run("upsert/"+tc.text, func(t *testing.T) {
			err := s.Upsert(UpsertItem{tc.id, tc.text, tc.emb})
			if err != nil {
				t.Fatalf("Upsert failed: %v", err)
			}
		})
	}

	for _, tc := range tests {
		t.Run("search/"+tc.text, func(t *testing.T) {
			res, err := s.Search(tc.emb, 1)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(res) != 1 || res[0].RawText != tc.text {
				t.Errorf("Search() = %v, want 1 result with RawText %q", res, tc.text)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	s, err := NewVectorStore(memoryPath, embeddingDim)
	if err != nil {
		t.Fatalf("NewVectorStore failed: %v", err)
	}

	emb := []float32{1, 0, 0}
	res, err := s.Search(emb, 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(res) != 0 {
		t.Errorf("Search() = %v, want 0 result", res)
	}
}

func TestCount(t *testing.T) {
	s, err := NewVectorStore(memoryPath, embeddingDim)
	if err != nil {
		t.Fatalf("NewVectorStore failed: %v", err)
	}

	var initialTests = []struct {
		id   int64
		text string
		emb  []float32
	}{
		{1, "Hello world!", []float32{1, 0, 0}},
		{2, "Bye world!", []float32{0.9, 0.1, -0.1}},
		{3, "Quack!", []float32{0.1, 0.1, -1}},
		{4, "Duck!", []float32{0.05, 0.15, -0.9}},
		{5, "The sky is blue!", []float32{0.45, 0.5, 0.6}},
	}

	for _, tc := range initialTests {
		t.Run("upsert/"+tc.text, func(t *testing.T) {
			err := s.Upsert(UpsertItem{tc.id, tc.text, tc.emb})
			if err != nil {
				t.Fatalf("Upsert failed: %v", err)
			}
		})
	}

	count, err := s.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != len(initialTests) {
		t.Errorf("Count() = %v, want %d result", count, len(initialTests))
	}

	emb := []float32{-1, 0, 0}
	rawText := "!dlrow olleH"

	err = s.Upsert(UpsertItem{1, rawText, emb})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	count, err = s.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != len(initialTests) {
		t.Errorf("Count() = %v, want %d result", count, len(initialTests))
	}
}
