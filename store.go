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

type VectorStore struct {
	db  *sql.DB
	mu  sync.Mutex
	dim uint16
}

func NewVectorStore(path string, dim uint16) (*VectorStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	var version string
	err = db.QueryRow(`SELECT vec_version()`).Scan(&version)
	if err != nil {
		db.Close()
		return nil, err
	}
	slog.Info("sqlite-vec loaded", "version", version)

	createItems :=
		`CREATE TABLE IF NOT EXISTS items (
			item_id 	INTEGER 	PRIMARY KEY,
			raw_text 	TEXT		NOT NULL
		)`
	_, err = db.Exec(createItems)
	if err != nil {
		db.Close()
		return nil, err
	}

	createVecItems := fmt.Sprintf(
		`CREATE VIRTUAL TABLE IF NOT EXISTS vec_items
			USING vec0(embedding float[%d] distance_metric=cosine)
			`,
		dim)
	_, err = db.Exec(createVecItems)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &VectorStore{
		db:  db,
		dim: dim,
	}, nil
}

type UpsertItem struct {
	ID        int64
	RawText   string
	Embedding []float32
}

func (s *VectorStore) Upsert(item UpsertItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	id := item.ID
	rawText := item.RawText
	emb := item.Embedding

	upsertItem := `INSERT INTO items (item_id, raw_text) VALUES (?, ?)
					ON CONFLICT (item_id)
					DO UPDATE SET raw_text = EXCLUDED.raw_text`
	_, err = tx.Exec(upsertItem, id, rawText)
	if err != nil {
		return err
	}

	embBytes, err := sqlite_vec.SerializeFloat32(emb)
	if err != nil {
		return err
	}

	deleteEmbedding := `DELETE FROM vec_items WHERE (rowid) = ?`
	_, err = tx.Exec(deleteEmbedding, id)
	if err != nil {
		return err
	}

	insertEmbedding := `INSERT INTO vec_items (rowid, embedding) VALUES (?, ?)`
	_, err = tx.Exec(insertEmbedding, id, embBytes)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *VectorStore) UpsertBatch(items []UpsertItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	for _, item := range items {
		id := item.ID
		rawText := item.RawText
		emb := item.Embedding

		upsertItem := `INSERT INTO items (item_id, raw_text) VALUES (?, ?)
						ON CONFLICT (item_id)
						DO UPDATE SET raw_text = EXCLUDED.raw_text`
		_, err = tx.Exec(upsertItem, id, rawText)
		if err != nil {
			return err
		}

		embBytes, err := sqlite_vec.SerializeFloat32(emb)
		if err != nil {
			return err
		}

		deleteEmbedding := `DELETE FROM vec_items WHERE (rowid) = ?`
		_, err = tx.Exec(deleteEmbedding, id)
		if err != nil {
			return err
		}

		insertEmbedding := `INSERT INTO vec_items (rowid, embedding) VALUES (?, ?)`
		_, err = tx.Exec(insertEmbedding, id, embBytes)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

type SearchResult struct {
	RawText         string
	SimilarityScore float64
}

func (s *VectorStore) Search(emb []float32, topK int) ([]SearchResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	embBytes, err := sqlite_vec.SerializeFloat32(emb)
	if err != nil {
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
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var result SearchResult
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
		return 0, err
	}

	return count, nil
}
