package download

import (
	"os"
	"testing"
)

func TestParseChecksumSum(t *testing.T) {
	content := `# comment
cdbb85b45f38f282f05764ca08648b5f92db99c75b2fb3848eb4a559f6553b48 *php-8.3.31-nts-Win32-vs16-x64.zip
cad69da3c81d8f35cc05ffe31dedd06a7292952a php-5.2.6-nts-Win32.zip
`
	m, err := ParseChecksumSum(content)
	if err != nil {
		t.Fatal(err)
	}
	if got := m["php-8.3.31-nts-Win32-vs16-x64.zip"]; got != "cdbb85b45f38f282f05764ca08648b5f92db99c75b2fb3848eb4a559f6553b48" {
		t.Fatalf("sha256 entry = %q", got)
	}
	if got := m["php-5.2.6-nts-Win32.zip"]; got != "cad69da3c81d8f35cc05ffe31dedd06a7292952a" {
		t.Fatalf("sha1 entry = %q", got)
	}
}

func TestVerifyChecksumSHA1(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/test.bin"
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	// sha1("hello")
	expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
	if err := VerifyChecksum(path, expected, AlgoSHA1); err != nil {
		t.Fatal(err)
	}
}
