package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"net/http"

	"k8s.io/klog/v2"

	"translog-nix/src/logs"
	"translog-nix/src/narinfo"
)

func fetchUpStream(url, address string) narinfo.NarInfo {
	resp, _ := http.Get(fmt.Sprintf("%s/%s.narinfo", url, address))
	defer resp.Body.Close()

	byteArray, _ := io.ReadAll(resp.Body)
	info, err := narinfo.Parse(string(byteArray))
	if err != nil {
		// ハンドル先の処理がカス
		// かと思ったが、log.Fatalはログ出力後にパニックするので問題無い
		log.Fatal(err)
	}
	return info
}

func scanInputText() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := scanner.Text()
	return text
}

func main() {
	const TARGET_CACHE = "https://cache.nixos.org"

	// 初回に旧チェックポイントが見つからない旨は情報ログなので、警告以上を表示する。
	klog.SetSlogLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))

	fmt.Printf("対象のストアパスを入力: ")
	inputAddr := scanInputText()
	info := fetchUpStream(TARGET_CACHE, inputAddr)
	// 追記する値をここで指定する。
	entry := logs.Entry{
		StorePathHash: inputAddr,
		NarHash:       info.NarHash,
		CacheURL:      TARGET_CACHE,
	}

	// 署名鍵を環境変数から読み込み、作業ディレクトリの ./logs に追記する。
	index, err := logs.Append(context.Background(), "./logs", os.Getenv("LOG_PRIVATE_KEY"), entry)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(index)
	fmt.Println("Success!")
}
