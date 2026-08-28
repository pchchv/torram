package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var downloadDir string

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   update.Message.Text,
	})
}

// Helper function for downloading a file by URL.
func downloadAndSaveFile(url, fileName string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed response, status: " + resp.Status)
	}

	// Forming the final path on the disk.
	finalPath := filepath.Join(downloadDir, fileName)
	out, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func startBot(token string) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if token == "" {
		log.Panic("Error: The TG_BOT_TOKEN environment variable is not set")
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		log.Panicf("bot initialization error: %v\n", err)
	}

	log.Println("Bot has been successfully started and is ready to go...")
	b.Start(ctx)
}
