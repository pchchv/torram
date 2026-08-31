package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	// 1. User uploaded a file.
	if update.Message.Document != nil {
		doc := update.Message.Document
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("File received: %s. Downloading...", doc.FileName),
		})

		// Getting the internal file path on Telegram servers.
		fileInfo, err := b.GetFile(ctx, &bot.GetFileParams{FileID: doc.FileID})
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Unable to retrieve information about the file."})
			return
		}

		// Generate the download URL (bot token required).
		downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), fileInfo.FilePath)
		// Download and save the file.
		if err = downloadAndSaveFile(downloadURL, doc.FileName); err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Error saving the file to the server."})
			log.Printf("Downloading error: %v\n", err)
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "The file has been successfully saved!"})
		return
	}

	// 2. The user posted a link in the text.
	text := strings.TrimSpace(update.Message.Text)
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "A link has been found. Trying to download...",
		})

		// Extract the file name from the URL (i.e., from the end of the link).
		fileName := filepath.Base(text)
		if fileName == "" || fileName == "." || fileName == "/" {
			fileName = "downloaded_file"
		}

		err := downloadAndSaveFile(text, fileName)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "The file could not be downloaded from the link. Please check if the link is working."})
			fmt.Printf("Error downloading from the link: %v\n", err)
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "The file has been successfully saved!"})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Please send me the torrent file or a direct link to it.",
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
	finalPath := filepath.Join(watchDir, fileName)
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
		bot.WithDefaultHandler(handleUpdate),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		log.Panicf("bot initialization error: %v\n", err)
	}

	log.Println("Bot has been successfully started and is ready to go...")
	b.Start(ctx)
}
