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

func cloudLog(message []byte, caption string) {
	parsed := make(map[string]interface{})
	if err := json.Unmarshal(message, &parsed); err != nil {
		return
	}
	logEntry := map[string]interface{}{
		"level":        "ERROR",
		"message":      caption,
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
	loadEnvFile()

	// Инициализируем базу данных
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Инициализируем бота
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		panic("TELEGRAM_TOKEN environment variable is not set")
	}

	var err error
	botInstance, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	// Создаем реестр обработчиков (автоматически регистрирует все обработчики)
	registry = handlers.NewRegistry(botInstance)
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

	var update tgbotapi.Update
	if err := json.Unmarshal(bodyData, &update); err != nil {
		return &Response{StatusCode: 400, Body: "Bad request"}, nil
	}

	if update.Message != nil {
		cloudLog(bodyData, "message")
		handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		cloudLog(bodyData, update.CallbackQuery.Data)
		handleCallbackQuery(update.CallbackQuery)
	}

	return &Response{StatusCode: 200, Body: "OK"}, nil
}

func handleMessage(message *tgbotapi.Message) {
	if message.IsCommand() {
		// Обрабатываем команду
		if err := registry.HandleCommand(message.Command(), message); err != nil {
			fmt.Printf("Failed to handle command %s: %v\n", message.Command(), err)
		}
	} else {
		// Обрабатываем текстовое сообщение
		if err := registry.HandleTextMessage(message); err != nil {
			fmt.Printf("Failed to handle text message: %v\n", err)
		}
	}
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Обрабатываем callback
	if err := registry.HandleCallback(callback.Data, callback); err != nil {
		fmt.Printf("Failed to handle callback %s: %v\n", callback.Data, err)
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
