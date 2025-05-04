package app

import (
	"context"
	"database/sql"
	"log/slog"

	"go.ytsaurus.tech/yt/go/ypath"
	"go.ytsaurus.tech/yt/go/yt"
)

type Storage interface {
	Get(key string) (*string, error)
	Set(key string, value string) error
}

type SQLiteStorage struct {
	logger   *slog.Logger
	filePath string
}

func NewSQLiteStorage(filePath string, logger *slog.Logger) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", filePath)
	defer db.Close()
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS storage (key string NOT NULL PRIMARY KEY, value string)`)
	if err != nil {
		return nil, err
	}
	return &SQLiteStorage{filePath: filePath, logger: logger}, nil
}

func (ts *SQLiteStorage) getDb() (*sql.DB, error) {
	return sql.Open("sqlite3", ts.filePath)
}

func (ts *SQLiteStorage) Get(key string) (*string, error) {
	ts.logger.Info("Getting value for key", slog.String("key", key))
	db, err := ts.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT value FROM storage WHERE key = ?`, key)
	if err != nil {
		return nil, err
	}
	res := rows.Next()
	if !res {
		return nil, nil
	}
	var value string
	err = rows.Scan(&value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (ts *SQLiteStorage) Set(key string, value string) error {
	ts.logger.Info("Setting key", slog.String("key", key))
	db, err := ts.getDb()
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO storage (key, value) VALUES (?, ?)`, key, value)
	if err != nil {
		return err
	}
	return nil
}

type YtStorage struct {
	logger   *slog.Logger
	ytc      yt.Client
	nodePath string
}

func NewYtStorage(ytc yt.Client, nodePath string, logger *slog.Logger) (*YtStorage, error) {
	_, err := ytc.CreateNode(
		context.Background(),
		ypath.Path(nodePath),
		yt.NodeDocument,
		&yt.CreateNodeOptions{
			Attributes: map[string]any{
				"value": map[string]any{},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return &YtStorage{ytc: ytc, nodePath: nodePath, logger: logger}, nil
}

func (s *YtStorage) Get(key string) (*string, error) {
	ctx := context.Background()
	var res string
	if err := s.ytc.GetNode(ctx, ypath.Path(s.nodePath).JoinChild("@value", key), &res, nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *YtStorage) Set(key string, value string) error {
	ctx := context.Background()
	if err := s.ytc.SetNode(ctx, ypath.Path(s.nodePath).JoinChild("@value", key), value, nil); err != nil {
		return err
	}
	return nil
}
