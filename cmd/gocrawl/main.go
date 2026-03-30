package main

import (
	"context"
	"fmt"
	"os"

	"github.com/suprt/gocrawl/internal/bootstrap"
)

func main() {

	app, err := bootstrap.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ctx := bootstrap.WaitForSignal()

	if app.Config.MaxDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, app.Config.MaxDuration)
		defer cancel()
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
