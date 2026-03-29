package naming

import (
	"strings"
	"testing"
)

func TestReadableNamer_Name(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantSuffix string // ожидаемое окончание имени (без хэша)
	}{
		{
			name:       "Простой URL",
			url:        "https://example.com",
			wantSuffix: "example",
		},
		{
			name:       "URL с путём",
			url:        "https://example.com/path/to/page",
			wantSuffix: "page",
		},
		{
			name:       "URL с www",
			url:        "https://www.example.com",
			wantSuffix: "example",
		},
		{
			name:       "Golang.org",
			url:        "https://golang.org",
			wantSuffix: "golang",
		},
		{
			name:       "URL с длинным путём",
			url:        "https://example.com/very/long/path/that/exceeds/limit",
			wantSuffix: "limit",
		},
	}

	namer := NewReadableNamer(50)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := namer.Name(tt.url)
			if !strings.HasPrefix(got, tt.wantSuffix) {
				t.Errorf("Name() = %q, want prefix %q", got, tt.wantSuffix)
			}
			if !strings.HasSuffix(got, ".html") {
				t.Errorf("Name() = %q, want suffix .html", got)
			}
		})
	}
}

func TestReadableNamer_NameWithExtension(t *testing.T) {
	namer := NewReadableNamer(50)

	tests := []struct {
		name string
		url  string
		ext  string
	}{
		{
			name: "HTML",
			url:  "https://example.com",
			ext:  ".html",
		},
		{
			name: "CSS",
			url:  "https://example.com/style.css",
			ext:  ".css",
		},
		{
			name: "JS",
			url:  "https://example.com/script.js",
			ext:  ".js",
		},
		{
			name: "JSON",
			url:  "https://api.example.com/data",
			ext:  ".json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := namer.NameWithExtension(tt.url, tt.ext)
			if !strings.HasSuffix(got, tt.ext) {
				t.Errorf("NameWithExtension() = %q, want suffix %q", got, tt.ext)
			}
		})
	}
}

func TestReadableNamer_Name_maxLen(t *testing.T) {
	namer := NewReadableNamer(10)
	url := "https://example.com/verylongfilename"

	got := namer.Name(url)
	// readable part (10) + "_" + hash (8) + ".html" (5) = 24
	if len(got) != 24 {
		t.Errorf("Name() len = %d, want 24", len(got))
	}
}

func TestReadableNamer_getReadablePart(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Хост без пути",
			url:  "https://example.com",
			want: "example",
		},
		{
			name: "Последний сегмент пути",
			url:  "https://example.com/path/to/page",
			want: "page",
		},
		{
			name: "URL с www",
			url:  "https://www.example.com",
			want: "example",
		},
		{
			name: "Пустой путь",
			url:  "https://example.com/",
			want: "example",
		},
	}

	namer := NewReadableNamer(50)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := namer.getReadablePart(tt.url)
			if got != tt.want {
				t.Errorf("getReadablePart() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_sanitize(t *testing.T) {
	tests := []struct {
		name string
		input string
		want  string
	}{
		{
			name:  "Валидное имя",
			input: "valid_name-123",
			want:  "valid_name-123",
		},
		{
			name:  "Спецсимволы",
			input: "invalid@name#",
			want:  "invalid_name",
		},
		{
			name:  "Пустая строка",
			input: "@@@@",
			want:  "file",
		},
		{
			name:  "Подчёркивания в конце",
			input: "name___",
			want:  "name",
		},
		{
			name:  "Множественные спецсимволы подряд",
			input: "hello@@@world",
			want:  "hello_world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize(tt.input)
			if got != tt.want {
				t.Errorf("sanitize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReadableNamer_SHA256_hash(t *testing.T) {
	namer := NewReadableNamer(50)
	
	// Один и тот же URL должен давать одинаковый хэш
	url := "https://example.com"
	name1 := namer.Name(url)
	name2 := namer.Name(url)
	
	if name1 != name2 {
		t.Errorf("Name() produced different hashes for same URL: %q vs %q", name1, name2)
	}
	
	// Разные URL должны давать разные хэши
	url2 := "https://example.org"
	name3 := namer.Name(url2)
	
	if name1 == name3 {
		t.Errorf("Name() produced same hash for different URLs")
	}
}
