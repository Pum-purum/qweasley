package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"qweasley/internal/database"
	"qweasley/internal/handlers"
	"qweasley/internal/repository"
	"qweasley/internal/utils"
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
	sessionMgr  *handlers.SessionManager
)

// cloudLog логирует сообщение с уровнем INFO в JSON формате
func cloudLog(message string) {
	logEntry := map[string]interface{}{
		"level":        "INFO",
		"message":      message,
		"stream_name ": "body",
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	log.Print(string(jsonData))
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

	// Инициализируем менеджер сессий
	sessionMgr = handlers.NewSessionManager()

	// Инициализируем реестр обработчиков
	initHandlers()
}

func initHandlers() {
	registry = handlers.NewRegistry()

	// Создаем обработчики команд
	startHandler := handlers.NewStartHandler(sessionMgr)
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

	cloudLog(string(bodyData))

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
			log.Printf("Failed to send message: %v", err)
		}
	} else {
		// Обработка текстовых ответов на вопросы
		responseText, keyboard := handleTextAnswer(message)

		msg := tgbotapi.NewMessage(message.Chat.ID, responseText)
		msg.ParseMode = "MarkdownV2"

		if keyboard != nil {
			msg.ReplyMarkup = keyboard
		}

		if keyboard == nil {
			msg.ReplyToMessageID = message.MessageID
		}

		if _, err := botInstance.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}
}

func handleStartCommand(message *tgbotapi.Message) {
	// Получаем обработчик старта
	startHandler := registry.GetStartHandler()
	if startHandler == nil {
		log.Printf("Start handler not found")
		return
	}

	// Пытаемся отправить фото с вопросом
	photoConfig, err := startHandler.HandleWithPhoto(message)
	if err == nil && photoConfig != nil {
		// Отправляем фото с вопросом
		if _, err := botInstance.Send(*photoConfig); err != nil {
			log.Printf("Failed to send photo: %v", err)
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
		log.Printf("Failed to send message: %v", err)
	}
}

func handleTextAnswer(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем сессию пользователя
	session := sessionMgr.GetSession(message.Chat.ID)
	if session == nil {
		return "Сессия истекла\\. Начните заново командой /start", nil
	}

	// Получаем вопрос из базы данных
	questionRepo := repository.NewQuestionRepository()
	question, err := questionRepo.GetByID(session.QuestionID)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get question for text answer", map[string]interface{}{
			"question_id": session.QuestionID,
			"chat_id":     message.Chat.ID,
			"user_answer": message.Text,
		})
		return "Произошла ошибка при проверке ответа", nil
	}

	// Проверяем ответ
	userAnswer := strings.ToLower(strings.TrimSpace(message.Text))
	correctAnswer := strings.ToLower(strings.TrimSpace(question.Answer))

	if userAnswer == correctAnswer {
		// Получаем чат пользователя
		chatRepo := repository.NewChatRepository()
		reactionRepo := repository.NewReactionRepository()

		chat, err := chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
		if err != nil {
			utils.LogErrorWithContext(err, "Failed to get or create chat for text answer", map[string]interface{}{
				"chat_id":     message.Chat.ID,
				"question_id": session.QuestionID,
			})
			return "Произошла ошибка при обработке ответа", nil
		}

		// Создаем реакцию "response"
		err = reactionRepo.CreateOrUpdateReaction(chat.ID, session.QuestionID, "response")
		if err != nil {
			utils.LogErrorWithContext(err, "Failed to create response reaction", map[string]interface{}{
				"chat_id":     chat.ID,
				"question_id": session.QuestionID,
			})
			return "Произошла ошибка при обработке ответа", nil
		}

		// Уменьшаем баланс
		err = chatRepo.DecreaseBalance(chat.ID)
		if err != nil {
			utils.LogErrorWithContext(err, "Failed to decrease balance for text answer", map[string]interface{}{
				"chat_id":     chat.ID,
				"question_id": session.QuestionID,
			})
			return "Произошла ошибка при обработке ответа", nil
		}

		// Проверяем наличие картинки ответа
		if question.AnswerPicture != nil && question.AnswerPicture.Path != nil {
			// Формируем URL картинки
			photoURL, err := getPictureURL(*question.AnswerPicture.Path)
			if err != nil {
				log.Printf("Failed to get picture URL: %v", err)
				// Продолжаем без картинки
			} else {

				// Формируем ответ
				caption := "*Это правильный ответ\\!*"
				if question.Comment != nil {
					caption += "\n\n" + escapeMarkdown(*question.Comment)
				}

				// Создаем конфигурацию фото
				photoConfig := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileURL(photoURL))
				photoConfig.Caption = caption
				photoConfig.ParseMode = "MarkdownV2"

				// Добавляем клавиатуру
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Продолжаем", "continue"),
						tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
					),
				)
				photoConfig.ReplyMarkup = keyboard

				// Отправляем фото
				if _, err := botInstance.Send(photoConfig); err != nil {
					log.Printf("Failed to send photo: %v", err)
				}

				// Возвращаем пустой ответ, так как фото уже отправлено
				return "", nil
			}
		}

		// Формируем ответ без фото
		responseText := "*Это правильный ответ\\!*"
		if question.Comment != nil {
			responseText += "\n\n" + escapeMarkdown(*question.Comment)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Продолжаем", "continue"),
				tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
			),
		)
		return responseText, &keyboard
	}

	return "Ответ неверный\\. Попробуйте еще раз", nil
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
		log.Printf("Failed to answer callback query: %v", err)
	}

	// Специальная обработка для fail callback с возможностью отправки фото
	if callback.Data == "fail" {
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
		log.Printf("Failed to send callback response: %v", err)
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
				log.Printf("Failed to send photo: %v", err)
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
		log.Printf("Failed to send message: %v", err)
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
