package main

import (
	"log"

	"github.com/suprt/gocrawl/internal/bootstrap"
)

func main() {

	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := bootstrap.WaitForSignal()

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}

}
