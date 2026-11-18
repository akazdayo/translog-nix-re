package main;

import (
	"fmt"
	"net/http"
	"io"
	"translog-nix/src/narinfo"
)

func main() {
	url := "https://cache.nixos.org"
	ia := "fkv51s1m2gsy6pz71nnyqyprbvfg7sd7"
	text_narinfo := get_upstream(url, ia)
	var info narinfo.NarInfo

	parsed, err := info.Parse(text_narinfo)

	if err != nil {
		return
	}

	fmt.Println("StorePath:", parsed.StorePath)
	fmt.Println("NarHash:", parsed.NarHash)
	fmt.Println("NarSize:", parsed.NarSize)
}


func get_upstream(url, inputAddressed string) string {
	resp, _ := http.Get(fmt.Sprintf("%s/%s.narinfo", url, inputAddressed))
	defer resp.Body.Close()

	byteArray, _ := io.ReadAll(resp.Body)
	return string(byteArray)
}

