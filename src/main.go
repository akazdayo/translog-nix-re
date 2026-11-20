package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"translog-nix/src/narinfo"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	registerRoutes(router, "./logs")
	log.Fatal(http.ListenAndServe(":3000", router))
}

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
