package google_docs_auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GetClient() (*http.Client, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	tokenFile := "token.json"
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}

		err = saveToken(tokenFile, token)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), token), nil
}

func getConfig() (*oauth2.Config, error) {
	credentialsFile, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, err
	}
	return google.ConfigFromJSON(credentialsFile, "https://www.googleapis.com/auth/spreadsheets")
}

func saveToken(path string, token *oauth2.Token) error {
	slog.Info(fmt.Sprintf("Saving credential file to: %s\n", path))
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
	return nil
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

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, err
	}
	return token, nil
}
