package db

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"github.com/naimbob95/sharejer/internal/model"
)

type DB struct {
	conn *sql.DB
}

func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	if err := migrate(conn); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func migrate(conn *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		filename TEXT NOT NULL,
		storage_path TEXT NOT NULL,
		size INTEGER NOT NULL,
		mime_type TEXT NOT NULL DEFAULT '',
		password_hash TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		download_count INTEGER NOT NULL DEFAULT 0
	);
	`
	_, err := conn.Exec(query)
	return err
}

func (d *DB) InsertFile(f *model.File) error {
	query := `
	INSERT INTO files (id, filename, storage_path, size, mime_type, password_hash, created_at, expires_at, download_count)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	var expiresAt *string
	if f.ExpiresAt != nil {
		s := f.ExpiresAt.UTC().Format(time.RFC3339Nano)
		expiresAt = &s
	}

	_, err := d.conn.Exec(query, f.ID, f.Filename, f.StoragePath, f.Size, f.MimeType,
		f.PasswordHash, f.CreatedAt.UTC().Format(time.RFC3339Nano), expiresAt, f.DownloadCount)
	return err
}

func (d *DB) GetFile(id string) (*model.File, error) {
	query := `SELECT id, filename, storage_path, size, mime_type, password_hash, created_at, expires_at, download_count FROM files WHERE id = ?;`
	row := d.conn.QueryRow(query, id)

	var f model.File
	var createdAt string
	var expiresAt *string

	err := row.Scan(&f.ID, &f.Filename, &f.StoragePath, &f.Size, &f.MimeType,
		&f.PasswordHash, &createdAt, &expiresAt, &f.DownloadCount)
	if err != nil {
		return nil, err
	}

	f.CreatedAt = parseTime(createdAt)
	if expiresAt != nil {
		t := parseTime(*expiresAt)
		f.ExpiresAt = &t
	}

	return &f, nil
}

// parseTime decodes a timestamp stored by InsertFile (RFC3339Nano, UTC),
// falling back to SQLite's CURRENT_TIMESTAMP format for any default rows.
// On failure it returns the zero time.
func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.UTC()
	}
	return time.Time{}
}

func (d *DB) DeleteFile(id string) error {
	_, err := d.conn.Exec(`DELETE FROM files WHERE id = ?;`, id)
	return err
}

// ExpiredFiles returns files whose expiry has passed, relative to now. Only the
// id and storage path are populated — enough for the sweeper to delete them.
// Comparison is done in Go (not SQL) so it's robust to the stored time format.
func (d *DB) ExpiredFiles(now time.Time) ([]model.File, error) {
	rows, err := d.conn.Query(`SELECT id, storage_path, expires_at FROM files WHERE expires_at IS NOT NULL;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []model.File
	for rows.Next() {
		var f model.File
		var expiresAt string
		if err := rows.Scan(&f.ID, &f.StoragePath, &expiresAt); err != nil {
			return nil, err
		}
		if now.After(parseTime(expiresAt)) {
			expired = append(expired, f)
		}
	}
	return expired, rows.Err()
}

func (d *DB) IncrementDownload(id string) error {
	query := `UPDATE files SET download_count = download_count + 1 WHERE id = ?;`
	_, err := d.conn.Exec(query, id)
	return err
}

func (d *DB) Close() error {
	return d.conn.Close()
}