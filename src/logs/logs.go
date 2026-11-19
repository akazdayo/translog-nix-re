package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/transparency-dev/tessera"
	"github.com/transparency-dev/tessera/storage/posix"
	"golang.org/x/mod/sumdb/note"
)

// Entry は上流の .narinfo の原文と取得元のキャッシュ URL を記録する。
type Entry struct {
	CacheURL string `json:"cache_url"`
	NarInfo  string `json:"narinfo"`
}

// Append は1件追記し、そのエントリを含む署名済みチェックポイントが公開された後にインデックスを返す。
func Append(ctx context.Context, dir, privateKey string, entry Entry) (uint64, error) {
	signer, err := note.NewSigner(privateKey)
	if err != nil {
		return 0, fmt.Errorf("load checkpoint signing key: %w", err)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return 0, fmt.Errorf("encode entry: %w", err)
	}

	// 追記処理を抜けるときに Tessera のバックグラウンド処理を止める。
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	driver, err := posix.New(ctx, posix.Config{Path: dir})
	if err != nil {
		return 0, fmt.Errorf("open log storage: %w", err)
	}

	// 1件ごとに終了するため、チェックポイントの公開間隔を短くする。
	options := tessera.NewAppendOptions().
		WithCheckpointSigner(signer).
		WithCheckpointInterval(100 * time.Millisecond)
	appender, shutdown, _, err := tessera.NewAppender(ctx, driver, options)
	if err != nil {
		return 0, fmt.Errorf("open appender: %w", err)
	}

	// Add が返す関数を呼び、ログ内でのインデックスが確定するまで待つ。
	index, err := appender.Add(ctx, tessera.NewEntry(data))()
	if err != nil {
		return 0, fmt.Errorf("append entry: %w", err)
	}
	// インデックス確定後、そのエントリを含むチェックポイントの公開を待つ。
	if err := shutdown(ctx); err != nil {
		return 0, fmt.Errorf("entry %d: waiting for checkpoint: %w", index.Index, err)
	}
	return index.Index, nil
}
