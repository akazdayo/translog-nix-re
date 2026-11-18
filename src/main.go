package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"k8s.io/klog/v2"

	"translog-nix/src/logs"
)

func main() {
	// 初回に旧チェックポイントが見つからない旨は情報ログなので、警告以上を表示する。
	klog.SetSlogLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	// 追記する値をここで指定する。
	entry := logs.Entry{
		StorePathHash: "fkv51s1m2gsy6pz71nnyqyprbvfg7sd7",
		NarHash:       "sha256:050rzmgxapsh3n8n6k68dxk6lr1683igd33bsix0vw0a2zgp1jix",
		CacheURL:      "https://cache.nixos.org",
	}

	// 署名鍵を環境変数から読み込み、作業ディレクトリの ./logs に追記する。
	index, err := logs.Append(context.Background(), "./logs", os.Getenv("LOG_PRIVATE_KEY"), entry)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(index)
}
