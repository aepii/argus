package argus

import (
	"context"
	"errors"

	"github.com/aepii/argus/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShardServer struct {
	pb.UnimplementedShardServiceServer
	store *VectorStore
}

func NewShardServer(store *VectorStore) (*ShardServer, error) {
	if store == nil {
		return nil, errors.New("store must be supplied")
	}

	return &ShardServer{
		store: store,
	}, nil
}

func (s *ShardServer) Upsert(ctx context.Context, req *pb.UpsertRequest) (*pb.UpsertResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request was canceled: %v", err)
	}

	pbItem := req.GetItem()
	if pbItem == nil {
		return nil, status.Error(codes.InvalidArgument, "item is required")
	}

	store := s.store
	dim := store.dim

	id := pbItem.Id
	rawText := pbItem.RawText
	emb := pbItem.Embedding

	if len(emb) != int(dim) {
		return nil, status.Errorf(codes.InvalidArgument,
			"embedding dimension mismatch: got %d, want %d", len(emb), dim)
	}

	err := store.Upsert(UpsertItem{ID: id, RawText: rawText, Embedding: emb})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upsert: %v", err)
	}

	return &pb.UpsertResponse{}, nil
}

func (s *ShardServer) UpsertBatch(ctx context.Context, req *pb.UpsertBatchRequest) (*pb.UpsertBatchResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request was canceled: %v", err)
	}

	pbItems := req.GetItems()
	if len(pbItems) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	store := s.store
	dim := store.dim

	items := make([]UpsertItem, len(pbItems))

	for i, pbItem := range pbItems {
		id := pbItem.Id
		rawText := pbItem.RawText
		emb := pbItem.Embedding

		if len(emb) != int(dim) {
			return nil, status.Errorf(codes.InvalidArgument,
				"embedding dimension mismatch: got %d, want %d", len(emb), dim)
		}

		items[i] = UpsertItem{ID: id, RawText: rawText, Embedding: emb}
	}

	err := store.UpsertBatch(items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upsert: %v", err)
	}

	return &pb.UpsertBatchResponse{}, nil
}

func (s *ShardServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request was canceled: %v", err)
	}

	store := s.store
	dim := store.dim
	emb := req.GetEmbedding()

	if len(emb) != int(dim) {
		return nil, status.Errorf(codes.InvalidArgument,
			"embedding dimension mismatch: got %d, want %d", len(emb), dim)
	}

	res, err := store.Search(emb, int(req.TopK))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search: %v", err)
	}

	pbSearchResults := make([]*pb.SearchResult, len(res))

	for i, r := range res {
		pbSearchResults[i] = &pb.SearchResult{RawText: r.RawText, SimilarityScore: r.SimilarityScore}
	}

	return &pb.SearchResponse{Results: pbSearchResults}, nil
}

func (s *ShardServer) Count(ctx context.Context, req *pb.CountRequest) (*pb.CountResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Errorf(codes.Canceled, "request was canceled: %v", err)
	}

	store := s.store

	count, err := store.Count()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count: %v", err)
	}

	return &pb.CountResponse{Count: int64(count)}, nil
}
