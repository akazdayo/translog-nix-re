package narinfo

import (
	"fmt"
	"strconv"
	"strings"
)

type NarInfo struct {
	StorePath   string
	URL         string
	Compression Compression
	FileHash    string
	FileSize    int
	NarHash     string
	NarSize     int
	References  string
	Deriver     string
	Sig         string
}

// Parse は .narinfo のテキストを読み取り、NarInfo を返す。
func Parse(text string) (NarInfo, error) {
	var n NarInfo
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return NarInfo{}, fmt.Errorf("invalid line: %q", line)
		}

		switch key {
		case "StorePath":
			n.StorePath = value

		case "URL":
			n.URL = value

		case "Compression":
			n.Compression = Compression(value)

		case "FileHash":
			n.FileHash = value

		case "FileSize":
			i, err := strconv.Atoi(value)
			if err != nil {
				return NarInfo{}, fmt.Errorf("FileSize is invalid type: %q", value)
			}
			n.FileSize = i

		case "NarHash":
			n.NarHash = value

		case "NarSize":
			i, err := strconv.Atoi(value)
			if err != nil {
				return NarInfo{}, fmt.Errorf("NarSize is invalid type: %q", value)
			}
			n.NarSize = i

		case "References":
			n.References = value

		case "Deriver":
			n.Deriver = value

		case "Sig":
			n.Sig = value
		}

	}

	return n, nil
}
