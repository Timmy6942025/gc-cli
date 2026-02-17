package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/timothy/gc-cli/internal/platform"
)

type CacheEntry struct {
	Kind      string          `json:"kind"`
	Key       string          `json:"key"`
	Data      json.RawMessage `json:"data"`
	ETag      string          `json:"etag,omitempty"`
	UpdatedAt time.Time       `json:"updated_at"`
	Stale     bool            `json:"stale"`
}

type ParityRow struct {
	Feature   string    `json:"feature"`
	Status    string    `json:"status"`
	Notes     string    `json:"notes,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SQLiteStore struct {
	db *sql.DB
}

func OpenSQLite() (*SQLiteStore, error) {
	stateDir, err := platform.StateDir()
	if err != nil {
		return nil, err
	}
	if err := platform.EnsureDir(stateDir); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(stateDir, "cache.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	store := &SQLiteStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteStore) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS cache_entries (
			kind TEXT NOT NULL,
			cache_key TEXT NOT NULL,
			data BLOB NOT NULL,
			etag TEXT,
			updated_at INTEGER NOT NULL,
			stale INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY(kind, cache_key)
		);`,
		`CREATE TABLE IF NOT EXISTS sync_state (
			name TEXT PRIMARY KEY,
			last_refresh INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS parity_matrix (
			feature TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			notes TEXT,
			updated_at INTEGER NOT NULL
		);`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}
	return nil
}

func (s *SQLiteStore) UpsertCacheEntry(ctx context.Context, entry CacheEntry) error {
	if entry.UpdatedAt.IsZero() {
		entry.UpdatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO cache_entries(kind, cache_key, data, etag, updated_at, stale)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(kind, cache_key)
		 DO UPDATE SET data=excluded.data, etag=excluded.etag, updated_at=excluded.updated_at, stale=excluded.stale;`,
		entry.Kind,
		entry.Key,
		[]byte(entry.Data),
		entry.ETag,
		entry.UpdatedAt.Unix(),
		boolToInt(entry.Stale),
	)
	if err != nil {
		return fmt.Errorf("upsert cache entry: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetCacheEntry(ctx context.Context, kind, key string) (*CacheEntry, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT kind, cache_key, data, etag, updated_at, stale FROM cache_entries WHERE kind = ? AND cache_key = ?`,
		kind,
		key,
	)
	var entry CacheEntry
	var updated int64
	var data []byte
	var stale int
	if err := row.Scan(&entry.Kind, &entry.Key, &data, &entry.ETag, &updated, &stale); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get cache entry: %w", err)
	}
	entry.Data = append(entry.Data[:0], data...)
	entry.UpdatedAt = time.Unix(updated, 0).UTC()
	entry.Stale = stale == 1
	return &entry, nil
}

func (s *SQLiteStore) ListCacheEntriesByKind(ctx context.Context, kind string) ([]CacheEntry, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT kind, cache_key, data, etag, updated_at, stale FROM cache_entries WHERE kind = ? ORDER BY cache_key`,
		kind,
	)
	if err != nil {
		return nil, fmt.Errorf("list cache entries: %w", err)
	}
	defer rows.Close()

	entries := make([]CacheEntry, 0)
	for rows.Next() {
		var entry CacheEntry
		var updated int64
		var data []byte
		var stale int
		if err := rows.Scan(&entry.Kind, &entry.Key, &data, &entry.ETag, &updated, &stale); err != nil {
			return nil, fmt.Errorf("scan cache entry: %w", err)
		}
		entry.Data = append(entry.Data[:0], data...)
		entry.UpdatedAt = time.Unix(updated, 0).UTC()
		entry.Stale = stale == 1
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cache entries: %w", err)
	}
	return entries, nil
}

func (s *SQLiteStore) MarkStale(ctx context.Context, kind, key string) error {
	if key == "" {
		_, err := s.db.ExecContext(ctx, `UPDATE cache_entries SET stale = 1 WHERE kind = ?`, kind)
		if err != nil {
			return fmt.Errorf("mark stale entries: %w", err)
		}
		return nil
	}
	_, err := s.db.ExecContext(ctx, `UPDATE cache_entries SET stale = 1 WHERE kind = ? AND cache_key = ?`, kind, key)
	if err != nil {
		return fmt.Errorf("mark stale entry: %w", err)
	}
	return nil
}

func (s *SQLiteStore) SetLastRefresh(ctx context.Context, name string, at time.Time) error {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO sync_state(name, last_refresh) VALUES(?, ?)
		 ON CONFLICT(name)
		 DO UPDATE SET last_refresh=excluded.last_refresh;`,
		name,
		at.Unix(),
	)
	if err != nil {
		return fmt.Errorf("set last refresh: %w", err)
	}
	return nil
}

func (s *SQLiteStore) LastRefresh(ctx context.Context, name string) (time.Time, error) {
	var ts int64
	err := s.db.QueryRowContext(ctx, `SELECT last_refresh FROM sync_state WHERE name = ?`, name).Scan(&ts)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("get last refresh: %w", err)
	}
	return time.Unix(ts, 0).UTC(), nil
}

func (s *SQLiteStore) SetParityStatus(ctx context.Context, row ParityRow) error {
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO parity_matrix(feature, status, notes, updated_at) VALUES(?, ?, ?, ?)
		 ON CONFLICT(feature)
		 DO UPDATE SET status=excluded.status, notes=excluded.notes, updated_at=excluded.updated_at;`,
		row.Feature,
		row.Status,
		row.Notes,
		row.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("set parity status: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListParityStatus(ctx context.Context) ([]ParityRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT feature, status, notes, updated_at FROM parity_matrix ORDER BY feature`)
	if err != nil {
		return nil, fmt.Errorf("list parity status: %w", err)
	}
	defer rows.Close()

	out := make([]ParityRow, 0)
	for rows.Next() {
		var row ParityRow
		var ts int64
		if err := rows.Scan(&row.Feature, &row.Status, &row.Notes, &ts); err != nil {
			return nil, fmt.Errorf("scan parity row: %w", err)
		}
		row.UpdatedAt = time.Unix(ts, 0).UTC()
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parity rows: %w", err)
	}
	return out, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
