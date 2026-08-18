package main

import (
	"log"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("application failed: %s", err)
	}
}
