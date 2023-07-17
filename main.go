package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	ErrTgFailedToInitBot = errors.New("failed to init telegram bot")
	ErrTgMessageNotSent  = errors.New("message not sent")
)

var (
	BotToken    string
	BearerToken string
	ChatId      int64
	Port        int
)

type Message struct {
	Text  string `json:"text"`
	Image string `json:"image"` // Base64 encoded image
}

func init() {
	var err error
	BotToken = os.Getenv("BOT_TOKEN")
	BearerToken = os.Getenv("BEARER_TOKEN")
	ChatId, err = strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		panic("failed to parse CHAT_ID")
	}
	Port, err = strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic("failed to parse PORT")
	}

	if BotToken == "" || BearerToken == "" || ChatId == 0 || Port == 0 {
		fmt.Println("BOT_TOKEN, BEARER_TOKEN and CHAT_ID env vars must be set")
		os.Exit(1)
	}
}

func main() {
	http.HandleFunc("/", messageHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
	if err != nil {
		panic(err)
	}
}

func sendToTelegram(msg Message) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTgFailedToInitBot, err)
	}
	//
	textMsg := tgbotapi.NewMessage(ChatId, msg.Text)
	_, err = bot.Send(textMsg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTgMessageNotSent, err)
	}

	return nil
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("got request from %s", r.RemoteAddr)
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+BearerToken {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	log.Printf("got message: %s", msg.Text)

	err = sendToTelegram(msg)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
