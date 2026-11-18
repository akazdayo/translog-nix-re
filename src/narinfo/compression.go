package narinfo

type Compression string

const (
	CompressionNone     Compression = "none"
	CompressionXZ       Compression = "xz"
	CompressionBzip2    Compression = "bzip2"
	CompressionGzip     Compression = "gzip"
	CompressionZstd     Compression = "zstd"
	CompressionBrotli   Compression = "br"
	CompressionCompress Compression = "compress"
	CompressionGrzip    Compression = "grzip"
	CompressionLrzip    Compression = "lrzip"
	CompressionLz4      Compression = "lz4"
	CompressionLzip     Compression = "lzip"
	CompressionLzma     Compression = "lzma"
	CompressionLzop     Compression = "lzop"
)

func (c Compression) String() string {
	return string(c)
}
