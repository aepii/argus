package argus

import (
	"database/sql"
	"log/slog"
	"sync"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

type VectorStore struct {
	db *sql.DB
	mu sync.Mutex
}

func init() {
	sqlite_vec.Auto()
}

func NewVectorStore(path string) (*VectorStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		slog.Error("failed to open connection", "error", err)
		return nil, err
	}

	var version string
	err = db.QueryRow(`SELECT vec_version()`).Scan(&version)
	if err != nil {
		slog.Error("failed to prepare vec_version", "error", err)
		db.Close()
		return nil, err
	}

	slog.Info("sqlite-vec loaded", "version", version)

	return &VectorStore{
		db: db,
	}, nil
}
