package download

import (
	"bufio"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	AlgoSHA256 = "sha256"
	AlgoSHA1   = "sha1"
)

// ParseChecksumSum parses GNU checksum list lines such as:
//   abc... *php-8.3.31-nts-Win32-vs16-x64.zip
//   deadbeef php-5.6.0-Win32.zip
func ParseChecksumSum(content string) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := strings.ToLower(parts[0])
		name := strings.TrimPrefix(parts[len(parts)-1], "*")
		result[name] = hash
	}
	return result, scanner.Err()
}

// ParseSHA256Sum is an alias for ParseChecksumSum.
func ParseSHA256Sum(content string) (map[string]string, error) {
	return ParseChecksumSum(content)
}

func hashFile(path string, algo string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var h hashWriter
	switch algo {
	case AlgoSHA256, "":
		h = sha256.New()
	case AlgoSHA1:
		h = sha1.New()
	default:
		return "", fmt.Errorf("unsupported checksum algorithm %q", algo)
	}

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type hashWriter interface {
	io.Writer
	Sum(b []byte) []byte
}

func SHA256File(path string) (string, error) {
	return hashFile(path, AlgoSHA256)
}

func SHA1File(path string) (string, error) {
	return hashFile(path, AlgoSHA1)
}

func VerifyChecksum(path, expected, algo string) error {
	actual, err := hashFile(path, algo)
	if err != nil {
		return err
	}
	expected = strings.ToLower(strings.TrimSpace(expected))
	if actual != expected {
		return fmt.Errorf("%s checksum mismatch: expected %s got %s", algo, expected, actual)
	}
	return nil
}

func VerifyFile(path, expected string) error {
	return VerifyChecksum(path, expected, AlgoSHA256)
}
