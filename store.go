package argus

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	sqlite_vec.Auto()
}

// Vector store
type VectorStore struct {
	db  *sql.DB
	mu  sync.Mutex
	dim uint16
}

func NewVectorStore(path string, dim uint16) (*VectorStore, error) {
	// Open a connection
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		slog.Error("failed to open connection", "error", err)
		return nil, err
	}

	// Query vec_version
	var version string
	err = db.QueryRow(`SELECT vec_version()`).Scan(&version)
	if err != nil {
		slog.Error("failed to query vec_version", "error", err)
		db.Close()
		return nil, err
	}
	slog.Info("sqlite-vec loaded", "version", version)

	// Create items table
	createItems :=
		`CREATE TABLE IF NOT EXISTS items (
			item_id 	INTEGER 	PRIMARY KEY,
			raw_text 	TEXT		NOT NULL
		)`
	_, err = db.Exec(createItems)
	if err != nil {
		slog.Error("failed to create items table", "error", err)
		db.Close()
		return nil, err
	}

	// Create vec_items virtual table
	createVecItems := fmt.Sprintf(
		`CREATE VIRTUAL TABLE IF NOT EXISTS vec_items
			USING vec0(embedding float[%d] distance_metric=cosine)
			`,
		dim)
	_, err = db.Exec(createVecItems)
	if err != nil {
		slog.Error("failed to create vec_items virtual table", "error", err)
		db.Close()
		return nil, err
	}

	// Return VectorStore
	return &VectorStore{
		db:  db,
		dim: dim,
	}, nil
}

func (s *VectorStore) Upsert(id int64, text string, emb []float32) error {
	// Hold lock
	s.mu.Lock()
	defer s.mu.Unlock()

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		slog.Error("failed to start transaction", "error", err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	// Upsert item
	upsertItem := `INSERT INTO items (item_id, raw_text) VALUES (?, ?)
					ON CONFLICT (item_id)
					DO UPDATE SET raw_text = EXCLUDED.raw_text`
	_, err = tx.Exec(upsertItem, id, text)
	if err != nil {
		slog.Error("failed to upsert item", "error", err, "id", id, "text", text)
		return err
	}

	// Upsert embedding
	embBytes, err := sqlite_vec.SerializeFloat32(emb)
	if err != nil {
		slog.Error("failed to serialize embedding", "error", err, "emb", emb)
		return err
	}

	deleteEmbedding := `DELETE FROM vec_items WHERE (rowid) = ?`
	_, err = tx.Exec(deleteEmbedding, id)
	if err != nil {
		slog.Error("failed to delete embedding", "error", err, "id", id)
		return err
	}

	insertEmbedding := `INSERT INTO vec_items (rowid, embedding) VALUES (?, ?)`
	_, err = tx.Exec(insertEmbedding, id, embBytes)
	if err != nil {
		slog.Error("failed to upsert vec_item", "error", err, "id", id, "emb", emb)
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// Search result
type Result struct {
	RawText         string
	SimilarityScore float64
}

func (s *VectorStore) Search(emb []float32, topK int) ([]Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Search for embedding
	embBytes, err := sqlite_vec.SerializeFloat32(emb)
	if err != nil {
		slog.Error("failed to serialize embedding", "error", err, "emb", emb)
		return nil, err
	}

	searchEmbedding := `SELECT i.raw_text, 1 - v.distance AS similarity
						FROM (
							SELECT rowid, distance 
							FROM vec_items
							WHERE embedding
							MATCH ?
							LIMIT ?
						) v
						JOIN items i on i.item_id = v.rowid
						ORDER BY v.distance`
	rows, err := s.db.Query(searchEmbedding, embBytes, topK)
	if err != nil {
		slog.Error("failed to search", "error", err, "emb", emb)
		return nil, err
	}
	defer rows.Close()

	// Aggregate results
	var results []Result
	for rows.Next() {
		var result Result
		if err := rows.Scan(&result.RawText, &result.SimilarityScore); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *VectorStore) Count() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	countItems := `SELECT COUNT(*) FROM items`
	err := s.db.QueryRow(countItems).Scan(&count)
	if err != nil {
		slog.Error("failed to count items", "error", err)
		return 0, err
	}

	return count, nil
}
