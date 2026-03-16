package parser

import (
	"io"
	"log"
	"os"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantLen     int
		wantErr     bool
		setupFile   bool
		fileContent string
	}{
		{
			name: "URLs из аргументов",
			cfg: Config{
				URLs:        []string{"https://example.com", "https://golang.org"},
				SkipInvalid: true,
				Deduplicate: true,
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "Дубликаты удаляются",
			cfg: Config{
				URLs:        []string{"https://example.com", "https://example.com"},
				SkipInvalid: true,
				Deduplicate: true,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Невалидный URL с SkipInvalid=true",
			cfg: Config{
				URLs:        []string{"https://example.com", "", ""},
				SkipInvalid: true,
				Deduplicate: true,
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "Пустой список URL",
			cfg: Config{
				URLs:        []string{},
				SkipInvalid: true,
				Deduplicate: true,
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "URLs из файла",
			cfg: Config{
				FilePath:    "test_urls.txt",
				SkipInvalid: true,
				Deduplicate: true,
			},
			wantLen:   4,
			wantErr:   false,
			setupFile: true,
			fileContent: `https://example.com
https://golang.org
https://google.com
# комментарий
невалидный
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				if err := os.WriteFile(tt.cfg.FilePath, []byte(tt.fileContent), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				defer os.Remove(tt.cfg.FilePath)
			}

			p := New(nil)
			got, err := p.Parse(tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("Parse() got len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestParser_normalize(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{
			name:   "HTTPS без схемы",
			rawURL: "example.com",
			want:   "https://example.com",
		},
		{
			name:   "Полный URL",
			rawURL: "https://example.com/path",
			want:   "https://example.com/path",
		},
		{
			name:   "URL с фрагментом (удаляется)",
			rawURL: "https://example.com#fragment",
			want:   "https://example.com",
		},
		{
			name:   "URL с заглавными буквами в хосте",
			rawURL: "https://EXAMPLE.com",
			want:   "https://example.com",
		},
		{
			name:    "Невалидный URL",
			rawURL:  "",
			wantErr: true,
		},
	}

	p := New(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.normalize(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalize() got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParser_deduplicate(t *testing.T) {
	p := New(nil)
	p.logger = log.New(io.Discard, "", 0) // отключаем логирование

	input := []string{
		"https://example.com",
		"https://golang.org",
		"https://example.com",
		"https://google.com",
		"https://golang.org",
	}

	got := p.deduplicate(input)

	if len(got) != 3 {
		t.Errorf("deduplicate() got len = %d, want 3", len(got))
	}

	seen := make(map[string]bool)
	for _, url := range got {
		if seen[url] {
			t.Errorf("deduplicate() found duplicate: %s", url)
		}
		seen[url] = true
	}
}

func TestParser_parseFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "parser_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `https://example.com
https://golang.org
# это комментарий
https://google.com

невалидный url
https://final.com
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	p := New(nil)
	got, err := p.parseFile(tmpFile.Name())

	if err != nil {
		t.Errorf("parseFile() error = %v", err)
	}

	// Ожидаем 5 URL (комментарии и пустые строки пропускаются)
	if len(got) != 5 {
		t.Errorf("parseFile() got len = %d, want 5", len(got))
	}
}
