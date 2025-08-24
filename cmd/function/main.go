package main

import (
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"os"
	"qweasley/internal/database"
	"qweasley/internal/handlers"
	"strings"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

type YandexCloudRequest struct {
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

var (
	botInstance *tgbotapi.BotAPI
	registry    *handlers.Registry
)

// cloudLog логирует сообщение без использования сторонних библиотек
func cloudLog(message []byte) {
	parsed := make(map[string]interface{})
	if err := json.Unmarshal(message, &parsed); err != nil {
		return
	}
	logEntry := map[string]interface{}{
		"level":        "ERROR",
		"message":      "body",
		"stream_name":  "body",
		"json-payload": parsed,
	}
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	// Выводим в stdout в JSON формате
	fmt.Println(string(jsonData))
}

func init() {
	if os.Getenv("LOCAL_TEST") == "true" {
		loadEnvFile()
	}

	// Инициализируем базу данных
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
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

	// Инициализируем реестр обработчиков
	initHandlers()
}

func initHandlers() {
	registry = handlers.NewRegistry()

	// Создаем обработчики команд
	startHandler := handlers.NewStartHandler()
	balanceHandler := handlers.NewBalanceHandler()
	rulesHandler := handlers.NewRulesHandler()
	feedbackHandler := handlers.NewFeedbackHandler()
	proposalHandler := handlers.NewProposalHandler()

	// Регистрируем команды
	registry.RegisterCommand(startHandler)
	registry.RegisterCommand(balanceHandler)
	registry.RegisterCommand(rulesHandler)
	registry.RegisterCommand(feedbackHandler)
	registry.RegisterCommand(proposalHandler)

	// Создаем и регистрируем callback обработчики
	registry.RegisterCallback(handlers.NewSkipCallback(startHandler))
	registry.RegisterCallback(handlers.NewFailCallback(startHandler))
	registry.RegisterCallback(handlers.NewContinueCallback(startHandler))
	registry.RegisterCallback(handlers.NewFinishCallback(startHandler))
}

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
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
			os.Setenv(key, value)
		}
	}
}

func Handler(ctx context.Context, request json.RawMessage) (*Response, error) {
	var bodyData []byte

	// Пробуем распарсить как обертку Yandex Cloud
	var cloudRequest YandexCloudRequest
	if err := json.Unmarshal(request, &cloudRequest); err != nil {
		bodyData = []byte(request)
	} else {
		// Извлекаем тело из обертки Yandex Cloud
		bodyData = []byte(cloudRequest.Body)
	}

	if len(bodyData) == 0 {
		return &Response{StatusCode: 400, Body: "Empty body"}, nil
	}

	cloudLog(bodyData)

	var update tgbotapi.Update
	if err := json.Unmarshal(bodyData, &update); err != nil {
		return &Response{StatusCode: 400, Body: "Bad request"}, nil
	}

	if update.Message != nil {
		handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery)
	}
	return &Response{StatusCode: 200, Body: "OK"}, nil
}

func handleMessage(message *tgbotapi.Message) {
	if message.IsCommand() && message.Command() == "start" {
		// Специальная обработка для команды /start с возможностью отправки фото
		handleStartCommand(message)
	} else if message.IsCommand() {
		// Обработка других команд
		responseText, keyboard := registry.HandleCommand(message.Command(), message)

		msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
		msg.ParseMode = "MarkdownV2"

		if keyboard != nil {
			msg.ReplyMarkup = keyboard
		}

		if _, err := botInstance.Send(msg); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	} else {
		// Обработка текстовых ответов на вопросы
		responseText, keyboard := registry.HandleTextMessage(message)

		msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
		msg.ParseMode = "MarkdownV2"

		if keyboard != nil {
			msg.ReplyMarkup = keyboard
		}

		if keyboard == nil {
			msg.ReplyToMessageID = message.MessageID
		}

		if _, err := botInstance.Send(msg); err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	}
}

func handleStartCommand(message *tgbotapi.Message) {
	// Получаем обработчик старта
	startHandler := registry.GetStartHandler()
	if startHandler == nil {
		fmt.Printf("Start handler not found\n")
		return
	}

	// Пытаемся отправить фото с вопросом
	photoConfig, err := startHandler.HandleWithPhoto(message)
	if err == nil && photoConfig != nil {
		// Отправляем фото с вопросом
		if _, err := botInstance.Send(*photoConfig); err != nil {
			fmt.Printf("Failed to send photo: %v\n", err)
		}
		return
	}

	// Если фото нет или произошла ошибка, отправляем обычное сообщение
	responseText, keyboard := startHandler.Handle(message)

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	msg.ParseMode = "MarkdownV2"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	if _, err := botInstance.Send(msg); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}
}

// getPictureURL формирует URL картинки
func getPictureURL(path string) (string, error) {
	endpoint := os.Getenv("AWS_S3_ENTRYPOINT")
	bucket := os.Getenv("AWS_S3_BUCKET")

	if endpoint == "" {
		return "", fmt.Errorf("AWS_S3_ENTRYPOINT environment variable is not set")
	}

	if bucket == "" {
		return "", fmt.Errorf("AWS_S3_BUCKET environment variable is not set")
	}

	return endpoint + "/" + bucket + "/" + path, nil
}

// escapeMarkdown экранирует специальные символы для Markdown
func escapeMarkdown(text string) string {
	specialChars := []string{"?", "!", "_", "*", "[", "]", "(", ")", "~", "`", ">", "<", "&", "#", "+", "-", "=", "|", "{", "}", "."}

	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}

	return text
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Отвечаем на callback query
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := botInstance.Send(callbackConfig); err != nil {
		fmt.Printf("Failed to answer callback query: %v\n", err)
	}

	// Специальная обработка для fail callback с возможностью отправки фото
	if strings.HasPrefix(callback.Data, "fail:") {
		handleFailCallback(callback)
		return
	}

	responseText, keyboard := registry.HandleCallback(callback.Data, callback)

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
	msg.ParseMode = "MarkdownV2"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	if _, err := botInstance.Send(msg); err != nil {
		fmt.Printf("Failed to send callback response: %v\n", err)
	}
}

func handleFailCallback(callback *tgbotapi.CallbackQuery) {
	// Получаем обработчик fail как конкретный тип
	if failHandler, ok := registry.CallbackHandlers["fail"].(*handlers.FailCallback); ok {
		// Пытаемся отправить фото с ответом
		photoConfig, err := failHandler.HandleWithPhoto(callback)
		if err == nil && photoConfig != nil {
			// Отправляем фото с ответом
			if _, err := botInstance.Send(*photoConfig); err != nil {
				fmt.Printf("Failed to send photo: %v\n", err)
			}
			return
		}
	}

	// Если фото нет или произошла ошибка, отправляем обычное сообщение
	responseText, keyboard := registry.HandleCallback("fail", callback)

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
	msg.ParseMode = "MarkdownV2"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	if _, err := botInstance.Send(msg); err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}
}

func main() {
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

		response, err := Handler(r.Context(), json.RawMessage(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(response.StatusCode)
		if bodyStr, ok := response.Body.(string); ok {
			w.Write([]byte(bodyStr))
		} else {
			json.NewEncoder(w).Encode(response.Body)
		}
	})

	fmt.Printf("Local server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
