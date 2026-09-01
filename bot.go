package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// BotHandler encapsulates dependencies for handling updates
type BotHandler struct {
	cfg *Config
}

func (h *BotHandler) downloadAndSaveFile(ctx context.Context, url, fileName string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed response, status: %s", resp.Status)
	}

	finalPath := filepath.Join(h.cfg.WatchDir, fileName)
	out, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (h *BotHandler) handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	// 1. Processing a document (.torrent file)
	if update.Message.Document != nil {
		doc := update.Message.Document
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("File received: %s. Downloading...", doc.FileName),
		})

		fileInfo, err := b.GetFile(ctx, &bot.GetFileParams{FileID: doc.FileID})
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Unable to retrieve information about the file."})
			return
		}

		downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), fileInfo.FilePath)
		if err = h.downloadAndSaveFile(ctx, downloadURL, doc.FileName); err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "Error saving the file to the server."})
			log.Printf("[Bot] Downloading error: %v", err)
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "The file has been successfully saved!"})
		return
	}

	// 2. Handling Links in the Text
	text := strings.TrimSpace(update.Message.Text)
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "A link has been found. Trying to download...",
		})

		fileName := filepath.Base(text)
		if fileName == "" || fileName == "." || fileName == "/" {
			fileName = "downloaded_file.torrent"
		}

		if err := h.downloadAndSaveFile(ctx, text, fileName); err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: "The file could not be downloaded from the link. Please check if the link is working."})
			log.Printf("[Bot] Error downloading from link: %v", err)
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

func runBot(ctx context.Context, cfg *Config) error {
	handler := &BotHandler{cfg: cfg}
	opts := []bot.Option{
		bot.WithDefaultHandler(handler.handleUpdate),
	}
	b, err := bot.New(cfg.BotToken, opts...)
	if err != nil {
		return fmt.Errorf("bot initialization: %w", err)
	}

	log.Println("[Bot] Telegram bot has been successfully started")
	b.Start(ctx)
	return nil
}
