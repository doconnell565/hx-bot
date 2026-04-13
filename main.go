package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/doconnell565/hx-bot/bot"
	"github.com/doconnell565/hx-bot/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	if err := b.Start(); err != nil {
		log.Fatalf("failed to start bot: %v", err)
	}
	defer b.Stop()

	log.Println("hx-bot is running. Press Ctrl+C to exit.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("shutting down...")
}
