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

type YandexCloudRequest struct {
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

var botInstance *tgbotapi.BotAPI

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
		switch message.Command() {
		case "start":
			responseText, keyboard = handleStartCommand(message)
		case "balance":
			responseText = handleBalanceCommand(message)
		case "rules":
			responseText = handleRulesCommand(message)
		case "feedback":
			responseText = handleFeedbackCommand(message)
		case "proposal":
			responseText = handleProposalCommand(message)
		default:
			responseText = "Неизвестная команда. Доступные команды: /start, /balance, /rules, /feedback, /proposal"
		}
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

func handleStartCommand(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Проверить баланс пользователя
	// TODO: Получить случайный вопрос из базы
	// TODO: Создать пользователя если не существует (30 монет)

	// Заглушка - показываем пример вопроса
	questionText := "*Вопрос:*\n\nКакая планета ближайшая к Солнцу?"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip"),
			tgbotapi.NewInlineKeyboardButtonData("Показать ответ", "fail"),
			tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
		),
	)

	return questionText, &keyboard
}

func handleBalanceCommand(message *tgbotapi.Message) string {
	// TODO: Получить баланс из базы данных
	balance := 30 // Заглушка

	return fmt.Sprintf("*Ваш баланс: %d монет\\.*\n\nПополнить баланс вы можете, предложив свой вопрос через соответствующую команду меню\\. В случае, если вопрос пройдет модерацию, он будет опубликован в боте и ваш счет будет пополнен на 10 монет\\. Если вы готовы приобрести монеты за деньги по курсу 1 монета \\= 10 рублей, свяжитесь с администрацией через команду \\/feedback", balance)
}

func handleRulesCommand(message *tgbotapi.Message) string {
	return "*Правила*\n\n1\\. При первом контакте с ботом на ваш счет закидывается 30 монет\\.\n2\\. За каждый верно отвеченный вопрос со счета снимается 1 монета\\.\n3\\. Ответом является одно слово на русском языке в именительном падеже единственного числа, если в вопросе не указано иное\\.\n4\\. Если ответом является калька с иностранного языка, имеющая несколько вариантов написания, то правильным будет тот, который указан в Википедии\\.\n5\\. Регистр букв в ответе не имеет значения\\.\n6\\. За каждое нажатие кнопки Показать ответ со счета снимается 1 монета\\.\n7\\. Счет привязан не к пользователю, а к чату\\.\n8\\. Монеты со счета нельзя вернуть\\, но можно отдать другому чату\\, для этого напишите в форму обратной связи\\.\n9\\. Бот поставляется \"как есть\"\\. Администрация не несет ответственности за любые негативные последствия, прямо или косвенно вызванные использованием бота\\."
}

func handleFeedbackCommand(message *tgbotapi.Message) string {
	// TODO: Реализовать форму обратной связи
	return "Напишите ваше сообщение администрации\\. Мы обязательно его прочитаем и ответим\\!"
}

func handleProposalCommand(message *tgbotapi.Message) string {
	// TODO: Реализовать форму предложения вопроса
	return "Предложите свой вопрос для квиза\\! Формат:\n\n*Вопрос:* Ваш вопрос\n*Ответ:* Правильный ответ\n*Комментарий:* Дополнительная информация \\(необязательно\\)"
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

	var responseText string
	var keyboard *tgbotapi.InlineKeyboardMarkup

	switch callback.Data {
	case "skip":

		// TODO: Пропустить вопрос, списать монету, показать следующий
		responseText, keyboard = handleStartCommand(&tgbotapi.Message{
			From: callback.From,
			Chat: callback.Message.Chat,
		})
	case "fail":
		// TODO: Показать правильный ответ, списать монету
		keyboard = &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("Точно!", "continue"),
					tgbotapi.NewInlineKeyboardButtonData("Ладно, хватит", "finish"),
				},
			},
		}
		responseText = "*Правильный ответ:*\nМеркурий\n\nМеркурий \\- самая близкая к Солнцу планета Солнечной системы\\."
	case "continue":
		// TODO: Показать следующий вопрос
		responseText, keyboard = handleStartCommand(&tgbotapi.Message{
			From: callback.From,
			Chat: callback.Message.Chat,
		})
	case "finish":
		responseText = "Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!"
	default:
		responseText = fmt.Sprintf("Получен callback: %s", callback.Data)
	}

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
