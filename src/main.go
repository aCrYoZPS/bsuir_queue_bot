package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	iis_api "github.com/aCrYoZPS/bsuir_queue_bot/src/iis_api"
	bot "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/bot"
	update_handlers "github.com/aCrYoZPS/bsuir_queue_bot/src/telegram/update_handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/joho/godotenv"
)

var opts = &slog.HandlerOptions{
	AddSource: false,
	Level:     slog.LevelDebug,
	ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
		if attr.Key != slog.TimeKey {
			return attr
		}

		curTime := attr.Value.Time()

		attr.Value = slog.StringValue(curTime.Format(time.DateTime))
		return attr
	},
}
var logger = slog.New(slog.NewTextHandler(os.Stderr, opts))

func main() {
	slog.SetDefault(logger)
	err := godotenv.Load()
	if err != nil {
		FatalLog("Error loading .env file")
	}

	//Perhaps, should handle API limits
	_, err = iis_api.GetAllGroups()
	if err != nil {
		FatalLog(err.Error())
	}

	credentialsFile, err := os.ReadFile("credentials.json")
	if err != nil {
		FatalLog(fmt.Sprintf("Unable to read client secret file: %v", err))
	}
	config, err := google.ConfigFromJSON(credentialsFile, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		FatalLog(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}
	client := getClient(config)

	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		FatalLog(fmt.Sprintf("Unable to retrieve sheets client: %v", err.Error()))
	}
	//Change later to do smth. As of now I am going to shop, so, whatever
	print(srv)
	InitBotController()
}

func FatalLog(message string) {
	slog.Error(message)
	os.Exit(-1)
}

func saveToken(path string, token *oauth2.Token) {
	slog.Info(fmt.Sprintf("Saving credential file to: %s\n", path))
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		FatalLog(fmt.Sprintf("Unable to cache oauth token: %v", err))
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(filename string) (*oauth2.Token, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(file).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		FatalLog(fmt.Sprintf("Unable to read authorization code: %v", err))
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		FatalLog(fmt.Sprintf("Unable to retrieve token from web: %v", err))
	}
	return tok
}

func InitBotController() {
	bot_token := os.Getenv("BOT_TOKEN")

	bot_controller := bot.GetBotController()

	err := bot_controller.Init(bot_token, true)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	bot, err := bot_controller.GetBot()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(-1)
	}

	slog.Info(fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	update_handlers.InitCommands()
	u := tgbotapi.NewUpdate(0)
	u.AllowedUpdates = []string{"callback_query", "message"}
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Command() != "" {
				update_handlers.HandleCommands(&update, bot)
			}
		} else if update.CallbackQuery != nil {
			update_handlers.HandleCallbacks(&update, bot)
		}
	}
}
