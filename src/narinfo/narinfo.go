package narinfo

import (
	"slices"
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
	Sig         []string
	CA          string
}

// Equal は .narinfo の正規化で吸収する表記の差を除いて比較する。
func (n NarInfo) Equal(other NarInfo) bool {
	return n.StorePath == other.StorePath &&
		n.URL == other.URL &&
		n.Compression == other.Compression &&
		n.FileHash == other.FileHash &&
		n.FileSize == other.FileSize &&
		n.NarHash == other.NarHash &&
		n.NarSize == other.NarSize &&
		n.CA == other.CA &&
		n.normalizedDeriver() == other.normalizedDeriver() &&
		slices.Equal(n.normalizedReferences(), other.normalizedReferences()) &&
		slices.Equal(n.normalizedSignatures(), other.normalizedSignatures())
}

func (n NarInfo) normalizedDeriver() string {
	if n.Deriver == "unknown-deriver" {
		return ""
	}
	return n.Deriver
}

// Nix は参照と署名を集合として扱う。
func (n NarInfo) normalizedReferences() []string {
	references := strings.Fields(n.References)
	slices.Sort(references)
	return slices.Compact(references)
}

func (n NarInfo) normalizedSignatures() []string {
	signatures := slices.Clone(n.Sig)
	slices.Sort(signatures)
	return slices.Compact(signatures)
}
