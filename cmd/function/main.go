package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

var botInstance *tgbotapi.BotAPI

func init() {
	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN environment variable is required")
	}

	var err error
	botInstance, err = tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}
}

func Handler(ctx context.Context) (*Response, error) {
	// Читаем тело запроса из stdin
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		return &Response{
			StatusCode: 400,
			Body:       "Bad request",
		}, nil
	}

	// Парсим JSON обновление напрямую
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		log.Printf("Error decoding update: %v", err)
		return &Response{
			StatusCode: 400,
			Body:       "Bad request",
		}, nil
	}

	// Обрабатываем сообщение
	if update.Message != nil {
		handleMessage(update.Message)
	}

	return &Response{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}

func handleMessage(message *tgbotapi.Message) {
	var responseText string

	// Обработка команд
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			responseText = handleStartCommand(message)
		default:
			responseText = "Неизвестная команда. Используйте /start для начала работы."
		}
	} else {
		responseText = "Echo: " + message.Text
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	if !message.IsCommand() {
		msg.ReplyToMessageID = message.MessageID
	}

	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func handleStartCommand(message *tgbotapi.Message) string {
	userName := message.From.FirstName
	if userName == "" {
		userName = message.From.UserName
	}
	if userName == "" {
		userName = "друг"
	}

	log.Printf("User %s (%d) started the bot", userName, message.From.ID)

	return fmt.Sprintf("Привет, %s! 👋\n\nДобро пожаловать в наш бот!\n\nЯ могу:\n• Отвечать на ваши сообщения\n• Обрабатывать команды\n\nПросто напишите мне что-нибудь!", userName)
}