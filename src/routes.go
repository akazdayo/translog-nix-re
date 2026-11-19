package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/transparency-dev/tessera/api/layout"
)

// タイルのパスは種別（レベル番号または entries）、番号、任意の部分タイル幅で構成される。
// 例として、entries/000.p/1 は最初のバンドルの先頭1件を指す。
var tilePath = regexp.MustCompile(`^(entries|0|[1-9][0-9]?)/(x[0-9]{3}/)*[0-9]{3}(\.p/[1-9][0-9]{0,2})?$`)

// registerRoutes は外部の検証者が Tessera のログを読み取るためのルートを登録する。
func registerRoutes(router chi.Router, logDir string) {
	files := logFileHandlers{dir: logDir}

	router.Get("/", serveHome)
	// 実行後にファイルの差し替えも可能ということは確認した
	router.Get("/checkpoint", files.serveCheckpoint)
	router.Head("/checkpoint", files.serveCheckpoint)
	router.Get("/tile/*", files.serveTile)
	router.Head("/tile/*", files.serveTile)
}

func serveHome(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("Hello World!"))
}

type logFileHandlers struct {
	dir string
}

// serveCheckpoint は最新の署名済みチェックポイントを返す。
// 追記で内容が変わるので、キャッシュを再利用する前にサーバーへの確認を求める。
func (h logFileHandlers) serveCheckpoint(w http.ResponseWriter, r *http.Request) {
	serveLogFile(w, r, h.dir, layout.CheckpointPath, "text/plain; charset=utf-8", "no-cache")
}

// serveTile は検証に使うハッシュタイルとエントリバンドルを返す。
// 公開済みのタイルは同じ URL で内容が固定されるので、長期キャッシュを許可する。
func (h logFileHandlers) serveTile(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "*")
	if !tilePath.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	tileDir := filepath.Join(h.dir, "tile")
	serveLogFile(w, r, tileDir, name, "application/octet-stream", "public, max-age=31536000, immutable")
}

func serveLogFile(w http.ResponseWriter, r *http.Request, dir, name, contentType, cacheControl string) {
	// シンボリックリンクをたどる場合も、指定したディレクトリ内のファイルに限定する。
	file, err := os.OpenInRoot(dir, name)
	if errors.Is(err, os.ErrNotExist) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "failed to open log file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to stat log file", http.StatusInternalServerError)
		return
	}
	if !info.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", cacheControl)
	// チェックポイントは同じ秒に更新され得るので、更新時刻で 304 を返さない。
	http.ServeContent(w, r, info.Name(), time.Time{}, file)
}
