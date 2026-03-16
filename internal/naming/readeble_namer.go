package naming

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
)

type ReadableNamer struct {
	maxLen int
}

func NewReadableNamer(maxLen int) *ReadableNamer {
	if maxLen <= 0 {
		maxLen = 50
	}
	return &ReadableNamer{maxLen: maxLen}
}

// Name генерирует имя файла с расширением .html по умолчанию
func (n *ReadableNamer) Name(rawURL string) string {
	return n.NameWithExtension(rawURL, ".html")
}

// NameWithExtension генерирует имя файла с указанным расширением
func (n *ReadableNamer) NameWithExtension(rawURL string, ext string) string {
	readable := n.getReadablePart(rawURL)
	if len(readable) > n.maxLen {
		readable = readable[:n.maxLen]
	}

	readable = sanitize(readable)
	hash := md5.Sum([]byte(rawURL))
	shortHash := hex.EncodeToString(hash[:])[:8]

	if readable == "" {
		return fmt.Sprintf("%s%s", shortHash, ext)
	}
	return fmt.Sprintf("%s_%s%s", readable, shortHash, ext)
}

func (n *ReadableNamer) getReadablePart(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	path := strings.TrimSuffix(parsed.Path, "/")
	if path != "" && path != "/" {
		segments := strings.Split(path, "/")
		last := segments[len(segments)-1]
		if last != "" {
			return last
		}
	}

	host := strings.TrimPrefix(parsed.Hostname(), "www.")
	if host != "" {
		parts := strings.Split(host, ".")
		return parts[0]
	}
	return ""
}

func sanitize(name string) string {
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_"
	var result strings.Builder
	lastWasUnderscore := false
	for _, r := range name {
		if strings.ContainsRune(allowed, r) {
			result.WriteRune(r)
			lastWasUnderscore = false
		} else {
			if !lastWasUnderscore {
				result.WriteRune('_')
				lastWasUnderscore = true
			}
		}
	}
	name = strings.TrimRight(result.String(), "_")
	if name == "" {
		return "file"
	}
	return name
}
