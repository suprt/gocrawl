package main

import (
	"context"
	"log"

	"github.com/suprt/gocrawl/internal/bootstrap"
)

func main() {

	app, err := bootstrap.New()
	if err != nil {
		log.Fatal(err)
	}

	// Создаём контекст с обработкой сигналов
	ctx := bootstrap.WaitForSignal()

	// Оборачиваем в таймаут, если задан
	if app.Config.MaxDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, app.Config.MaxDuration)
		defer cancel()
	}

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}

}
