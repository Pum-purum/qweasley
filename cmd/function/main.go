package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

var botInstance *tgbotapi.BotAPI

func init() {
	// Загружаем .env файл при локальном запуске
	if os.Getenv("LOCAL_TEST") == "true" {
		loadEnvFile()
	}

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

// loadEnvFile загружает переменные из .env файла
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading .env file: %v", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			err := os.Setenv(key, value)
			if err != nil {
				return
			}
		}
	}
}

// YandexCloudRequest представляет обертку от Yandex Cloud
type YandexCloudRequest struct {
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func Handler(ctx context.Context, request json.RawMessage) (*Response, error) {
	log.Printf("=== HANDLER CALLED ===")

	var bodyData []byte

	// Сначала пробуем распарсить как обертку Yandex Cloud
	var cloudRequest YandexCloudRequest
	if err := json.Unmarshal(request, &cloudRequest); err != nil {
		log.Printf("Not Yandex Cloud wrapper, using direct data: %v", err)
		// Если не получилось, считаем что это прямые данные Telegram (локальное тестирование)
		bodyData = []byte(request)
	} else {
		// Успешно распарсили обертку, извлекаем тело
		log.Printf("Yandex Cloud request, method: %s", cloudRequest.HTTPMethod)
		log.Printf("Body from wrapper: %s", cloudRequest.Body)
		bodyData = []byte(cloudRequest.Body)
	}

	log.Printf("Final body length: %d", len(bodyData))
	log.Printf("Final body data: %s", string(bodyData))

	if len(bodyData) == 0 {
		log.Printf("Empty body received")
		return &Response{StatusCode: 400, Body: "Empty body"}, nil
	}

	var update tgbotapi.Update
	if err := json.Unmarshal(bodyData, &update); err != nil {
		log.Printf("Error parsing body: %v", err)
		log.Printf("Raw bytes: %v", bodyData)
		return &Response{StatusCode: 400, Body: "Bad request"}, nil
	}

	log.Printf("Parsed update: UpdateID=%d", update.UpdateID)

	// Обрабатываем разные типы обновлений
	if update.Message != nil {
		log.Printf("Message: %v", update.Message.Text)
		handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		log.Printf("Callback query: %v", update.CallbackQuery.Data)
		handleCallbackQuery(update.CallbackQuery)
	} else {
		log.Printf("Unknown update type: %+v", update)
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

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	log.Print("Hello!")

	// Обрабатываем данные callback'а
	var responseText string
	switch callback.Data {
	case "start":
		responseText = "Вы нажали кнопку 1!"
	case "finish":
		responseText = "Вы нажали кнопку 2!"
	default:
		responseText = fmt.Sprintf("Получен callback: %s", callback.Data)
	}

	// Отправляем ответное сообщение
	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Failed to send callback response: %v", err)
	}
}

// Локальный HTTP сервер для тестирования
func main() {
	// Проверяем, запущен ли локально
	if os.Getenv("LOCAL_TEST") == "true" {
		startLocalServer()
	}
}

func startLocalServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Вызываем Handler с телом запроса как json.RawMessage
		response, err := Handler(r.Context(), json.RawMessage(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Возвращаем ответ от Handler
		w.WriteHeader(response.StatusCode)
		if bodyStr, ok := response.Body.(string); ok {
			_, err := w.Write([]byte(bodyStr))
			if err != nil {
				return
			}
		} else {
			err := json.NewEncoder(w).Encode(response.Body)
			if err != nil {
				return
			}
		}
	})

	log.Printf("Local server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
