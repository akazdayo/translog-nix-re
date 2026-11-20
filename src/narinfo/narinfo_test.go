package narinfo

import "testing"

func TestStringNormalizesNarInfo(t *testing.T) {
	input := "Sig: z:signature\r\n" +
		"NarSize: 00042\r\n" +
		"References: z-reference  a-reference z-reference\r\n\r\n" +
		"NarHash: sha256:hash\r\n" +
		"StorePath: /nix/store/example\r\n" +
		"Sig: a:signature\r\n" +
		"Sig: z:signature\r\n" +
		"Deriver: unknown-deriver\r\n" +
		"FileSize: 00021\r\n" +
		"CA: fixed:r:sha256:hash\r\n" +
		"URL: nar/example.nar.xz\r\n" +
		"Compression: xz\r\n" +
		"FileHash: sha256:file"

	want := `StorePath: /nix/store/example
URL: nar/example.nar.xz
Compression: xz
FileHash: sha256:file
FileSize: 21
NarHash: sha256:hash
NarSize: 42
References: a-reference z-reference
Sig: a:signature
Sig: z:signature
CA: fixed:r:sha256:hash
`

	info, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := info.String()
	if got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if info.Sig[0] != "z:signature" {
		t.Fatalf("String() changed Sig: %v", info.Sig)
	}

	normalized, err := Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if normalized.String() != got {
		t.Fatal("serializing normalized NarInfo changed its output")
	}
	if !info.Equal(normalized) {
		t.Fatal("Equal rejected differences that normalization removes")
	}
	normalized.NarHash = "sha256:changed"
	if info.Equal(normalized) {
		t.Fatal("Equal ignored a changed NarHash")
	}
}

func TestStringWritesEmptyReferences(t *testing.T) {
	input := `StorePath: /nix/store/example
URL: nar/example.nar
Compression: none
NarHash: sha256:hash
NarSize: 42
`
	info, err := Parse(input + "References:\n")
	if err != nil {
		t.Fatal(err)
	}
	want := input + "References: \n"
	if got := info.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
