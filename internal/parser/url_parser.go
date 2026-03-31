package parser

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

type Parser struct {
	logger *log.Logger
}
type Config struct {
	FilePath    string
	URLs        []string
	SkipInvalid bool
	Deduplicate bool
}

func New(logger *log.Logger) *Parser {
	if logger == nil {
		logger = log.New(os.Stderr, "[parser] ", log.LstdFlags)
	}
	return &Parser{
		logger: logger,
	}
}

func (p *Parser) Parse(cfg Config) ([]string, error) {
	var allURLs []string

	if cfg.FilePath != "" {
		fileURLs, err := p.parseFile(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("parse file: %w", err)
		}
		allURLs = append(allURLs, fileURLs...)
	}

	allURLs = append(allURLs, cfg.URLs...)

	if len(allURLs) == 0 {
		return nil, fmt.Errorf("no URLs provided")
	}

	normalized, err := p.normalizeAll(allURLs, cfg.SkipInvalid)
	if err != nil {
		return nil, err
	}

	if cfg.Deduplicate {
		normalized = p.deduplicate(normalized)
	}

	return normalized, nil
}

func (p *Parser) parseFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			p.logger.Printf("file close error: %s", err)
		}
	}(file)

	var urls []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file at line %d: %w", lineNum, err)
	}

	p.logger.Printf("parsed %d URL(s) from file %s", len(urls), filePath)
	return urls, nil
}

func (p *Parser) normalizeAll(rawURLs []string, skipInvalid bool) ([]string, error) {
	var result []string
	var errorMessages []string

	for _, rawURL := range rawURLs {
		normalized, err := p.normalize(rawURL)
		if err != nil {
			msg := fmt.Sprintf("invalid URL %q: %v", rawURL, err)
			if skipInvalid {
				p.logger.Printf("Warning: %s\n", msg)
				errorMessages = append(errorMessages, msg)
				continue
			}
			return nil, fmt.Errorf("%s", msg)
		}

		result = append(result, normalized)
	}

	if skipInvalid && len(errorMessages) > 0 {
		p.logger.Printf("Skipped %d invalid URLs", len(errorMessages))
	}
	return result, nil
}

func (p *Parser) normalize(rawURL string) (string, error) {
	if strings.HasPrefix(rawURL, "javascript:") ||
		strings.HasPrefix(rawURL, "data:") ||
		strings.HasPrefix(rawURL, "file:") {
		return "", fmt.Errorf("invalid URL scheme")
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("parse url: missing host")
	}
	parsed.Fragment = ""

	parsed.Host = strings.ToLower(parsed.Host)

	return parsed.String(), nil
}

func (p *Parser) deduplicate(rawURLs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, rawURL := range rawURLs {
		if !seen[rawURL] {
			seen[rawURL] = true
			result = append(result, rawURL)
		}
	}
	if len(result) < len(rawURLs) {
		p.logger.Printf("Removed %d duplicate URLs", len(rawURLs)-len(result))
	}
	return result
}
