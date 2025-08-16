package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"qweasley/internal/handlers"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func init() {
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
	registry.RegisterCallback(handlers.NewFailCallback())
	registry.RegisterCallback(handlers.NewContinueCallback(startHandler))
	registry.RegisterCallback(handlers.NewFinishCallback())
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
		// Прямые данные Telegram (локальное тестирование)
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

	// Обрабатываем разные типы обновлений
	if update.Message != nil {
		handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery)
	}

	return &Response{StatusCode: 200, Body: "OK"}, nil
}

func handleMessage(message *tgbotapi.Message) {
	var responseText string
	var keyboard *tgbotapi.InlineKeyboardMarkup

	if message.IsCommand() {
		responseText, keyboard = registry.HandleCommand(message.Command(), message)
	} else {
		// Обработка текстовых ответов на вопросы
		responseText, keyboard = handleTextAnswer(message)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	msg.ParseMode = "MarkdownV2"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	if !message.IsCommand() && keyboard == nil {
		msg.ReplyToMessageID = message.MessageID
	}

	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func handleTextAnswer(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Проверить ответ на текущий вопрос
	// TODO: Сравнить с правильным ответом из базы

	// Заглушка - проверяем ответ "Меркурий"
	userAnswer := strings.ToLower(strings.TrimSpace(message.Text))
	correctAnswer := "меркурий"

	if userAnswer == correctAnswer {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Продолжаем", "continue"),
				tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
			),
		)
		return "*Это правильный ответ\\!*\n\nМеркурий \\- самая близкая к Солнцу планета Солнечной системы\\.", &keyboard
	}

	return "Ответ неверный\\. Попробуйте еще раз", nil
}

func handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Отвечаем на callback query
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := botInstance.Send(callbackConfig); err != nil {
		log.Printf("Failed to answer callback query: %v", err)
	}

	responseText, keyboard := registry.HandleCallback(callback.Data, callback)

	msg := tgbotapi.NewMessage(callback.Message.Chat.ID, responseText)
	msg.ParseMode = "MarkdownV2"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Failed to send callback response: %v", err)
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

	log.Printf("Local server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
