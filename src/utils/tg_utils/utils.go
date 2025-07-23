package tgutils

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var owners []string = nil

func SendMessageToOwners(msg tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) error {
	if owners == nil {
		owners = strings.Split(os.Getenv("OWNERS"), ",")
	}

	for _, owner := range owners {
		chatId, err := strconv.ParseInt(owner, 10, 64)
		if err != nil {
			return errors.Join(err, fmt.Errorf("invalid owner id value %s", owner))
		}

		msg.ChatID = chatId
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func SelectMaxSizedPhoto(sizes []tgbotapi.PhotoSize) string {
	maxSize := 0
	maxSizeId := ""
	for _, photo := range sizes {
		if photo.FileSize > maxSize {
			maxSizeId = photo.FileID
		}
	}
	return maxSizeId
}