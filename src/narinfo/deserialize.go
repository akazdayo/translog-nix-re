package narinfo

import (
	"strconv"
	"strings"
)

// String は、フィールド順と表記をそろえた .narinfo を返す。
func (n NarInfo) String() string {
	var text strings.Builder
	writeField := func(key, value string) {
		text.WriteString(key)
		text.WriteString(": ")
		text.WriteString(value)
		text.WriteByte('\n')
	}

	writeField("StorePath", n.StorePath)
	writeField("URL", n.URL)
	if n.Compression != "" {
		writeField("Compression", n.Compression.String())
	}
	if n.FileHash != "" {
		writeField("FileHash", n.FileHash)
	}
	if n.FileSize != 0 {
		writeField("FileSize", strconv.Itoa(n.FileSize))
	}
	writeField("NarHash", n.NarHash)
	writeField("NarSize", strconv.Itoa(n.NarSize))

	writeField("References", strings.Join(n.normalizedReferences(), " "))

	if deriver := n.normalizedDeriver(); deriver != "" {
		writeField("Deriver", deriver)
	}
	for _, signature := range n.normalizedSignatures() {
		writeField("Sig", signature)
	}
	if n.CA != "" {
		writeField("CA", n.CA)
	}

	return text.String()
}
