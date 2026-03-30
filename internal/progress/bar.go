package progress

import (
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

type ProgBar struct {
	bar *progressbar.ProgressBar
}

func New(max int, description string) *ProgBar {
	bar := progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(50),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionOnCompletion(func() {
			io.WriteString(os.Stdout, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	return &ProgBar{bar: bar}
}

func (b *ProgBar) Add(n int) error {
	err := b.bar.Add(n)
	if err != nil {
		return err
	}
	return nil
}

func (b *ProgBar) Finish() error {
	err := b.bar.Finish()
	if err != nil {
		return err
	}
	return nil
}

type NoopBar struct{}

func (b *NoopBar) Add(n int) error { return nil }
func (b *NoopBar) Finish() error   { return nil }
