package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"qweasley/internal/models"
	"qweasley/internal/repository"
	"strings"
	"time"
)

// TextHandler интерфейс для обработчиков текстовых сообщений
type TextHandler interface {
	Handle(message *tgbotapi.Message) error
}

// BaseHandler содержит общую логику для всех обработчиков
type BaseHandler struct {
	chatRepo     *repository.ChatRepository
	questionRepo *repository.QuestionRepository
	reactionRepo *repository.ReactionRepository
	bot          *tgbotapi.BotAPI
}

// NewBaseHandler создает новый базовый обработчик
func NewBaseHandler(bot *tgbotapi.BotAPI) *BaseHandler {
	return &BaseHandler{
		chatRepo:     repository.NewChatRepository(),
		questionRepo: repository.NewQuestionRepository(),
		reactionRepo: repository.NewReactionRepository(),
		bot:          bot,
	}
}

// GetOrCreateChat получает или создает чат пользователя
func (h *BaseHandler) GetOrCreateChat(telegramID int64, title *string) (*models.Chat, error) {
	return h.chatRepo.GetOrCreate(telegramID, title)
}

// CheckBalance проверяет баланс чата
func (h *BaseHandler) CheckBalance(chat *models.Chat) error {
	if chat.Balance <= 0 {
		return fmt.Errorf("insufficient balance")
	}
	return nil
}

// GetQuestionForChat получает вопрос для чата
func (h *BaseHandler) GetQuestionForChat(chat *models.Chat) (*models.Question, error) {
	return h.questionRepo.GetQuestion(chat, h.reactionRepo)
}

// SetWaitingAnswer устанавливает ожидание ответа
func (h *BaseHandler) SetWaitingAnswer(chatID uint, questionID uint, expiresIn time.Duration) error {
	return h.chatRepo.SetWaitingAnswer(chatID, questionID, expiresIn)
}

// ClearWaitingAnswer очищает ожидание ответа
func (h *BaseHandler) ClearWaitingAnswer(chatID uint) error {
	return h.chatRepo.ClearWaitingAnswer(chatID)
}

// SetWaitingFeedback устанавливает ожидание обратной связи
func (h *BaseHandler) SetWaitingFeedback(chatID uint, expiresIn time.Duration) error {
	return h.chatRepo.SetWaitingFeedback(chatID, expiresIn)
}

// ClearWaitingFeedback очищает ожидание обратной связи
func (h *BaseHandler) ClearWaitingFeedback(chatID uint) error {
	return h.chatRepo.ClearWaitingFeedback(chatID)
}

// DecreaseBalance уменьшает баланс
func (h *BaseHandler) DecreaseBalance(chatID uint) error {
	return h.chatRepo.DecreaseBalance(chatID)
}

// CreateQuestionKeyboard создает клавиатуру для вопроса
func (h *BaseHandler) CreateQuestionKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip"),
				tgbotapi.NewInlineKeyboardButtonData("Показать ответ", "fail"),
				tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
			},
		},
	}
}

// CreateContinueKeyboard создает клавиатуру для продолжения
func (h *BaseHandler) CreateContinueKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Точно!", "continue"),
				tgbotapi.NewInlineKeyboardButtonData("Ладно, хватит", "finish"),
			},
		},
	}
}

// GetPictureURL формирует URL картинки
func (h *BaseHandler) GetPictureURL(path string) (string, error) {
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

// EscapeMarkdown экранирует специальные символы для Markdown
func (h *BaseHandler) EscapeMarkdown(text string) string {
	specialChars := []string{"?", "!", "_", "*", "[", "]", "(", ")", "~", "`", ">", "<", "&", "#", "+", "-", "=", "|", "{", "}", "."}

	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}

	return text
}

// FormatQuestionText форматирует текст вопроса
func (h *BaseHandler) FormatQuestionText(question *models.Question) string {
	text := "*" + h.EscapeMarkdown(question.Text) + "*"

	// Добавляем рейтинг, если он есть (как в PHP-версии)
	if question.Rating != nil && *question.Rating > 0 {
		ratingText := fmt.Sprintf("На этот вопрос отвечают %d%% пользователей", *question.Rating)
		text += "\n\n_" + h.EscapeMarkdown(ratingText) + "_"
	}

	return text
}

// ProcessStartCommand обрабатывает общую логику команды start
func (h *BaseHandler) ProcessStartCommand(telegramID int64, title *string) (*models.Chat, *models.Question, error) {
	// Получаем или создаем чат пользователя
	chat, err := h.GetOrCreateChat(telegramID, title)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get or create chat: %v", err)
	}

	// Проверяем баланс
	if err := h.CheckBalance(chat); err != nil {
		return nil, nil, err
	}

	// Получаем вопрос для пользователя
	question, err := h.GetQuestionForChat(chat)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get question: %v", err)
	}

	// Если вопросов больше нет
	if question == nil {
		return nil, nil, fmt.Errorf("no questions available")
	}

	// Устанавливаем ожидание ответа на вопрос (30 минут)
	if err := h.SetWaitingAnswer(chat.ID, question.ID, 30*time.Minute); err != nil {
		return nil, nil, fmt.Errorf("failed to set waiting answer: %v", err)
	}

	return chat, question, nil
}

// ProcessTextResponse обрабатывает текстовый ответ на вопрос
func (h *BaseHandler) ProcessTextResponse(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup, string, error) {
	// Получаем или создаем чат пользователя
	chat, err := h.GetOrCreateChat(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to get or create chat: %v", err)
	}

	// Проверяем, ждет ли чат ответа на вопрос
	if !chat.IsWaitingAnswer() {
		// Чат не ждет ответа - игнорируем сообщение
		return "", nil, "", nil
	}

	// Получаем вопрос по ID из чата
	question, err := h.questionRepo.GetByID(*chat.LastQuestionID)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to get question: %v", err)
	}

	// Проверяем ответ
	userAnswer := strings.ToLower(strings.TrimSpace(message.Text))
	correctAnswer := strings.ToLower(strings.TrimSpace(question.Answer))

	if userAnswer == correctAnswer {
		// Обрабатываем правильный ответ
		err = h.ProcessUserReaction(chat.ID, question.ID, "response")
		if err != nil {
			return "", nil, "", fmt.Errorf("failed to process response reaction: %v", err)
		}

		// Формируем ответ
		responseText := "*Это правильный ответ\\!*"
		if question.Comment != nil {
			responseText += "\n\n" + h.EscapeMarkdown(*question.Comment)
		}

		keyboard := h.CreateContinueKeyboard()

		// Проверяем наличие картинки ответа
		var photoURL string
		if question.AnswerPicture != nil && question.AnswerPicture.Path != nil {
			photoURL, err = h.GetPictureURL(*question.AnswerPicture.Path)
			if err != nil {
				fmt.Printf("Failed to get answer picture URL: %v (path: %s)\n", err, *question.AnswerPicture.Path)
				// Если не удалось получить картинку, возвращаем пустой URL
				photoURL = ""
			}
		}

		return responseText, keyboard, photoURL, nil
	}

	return "Ответ неверный\\. Попробуйте еще раз", nil, "", nil
}

// GetNextQuestion получает следующий вопрос для чата
func (h *BaseHandler) GetNextQuestion(telegramID int64, title *string) (*models.Question, *tgbotapi.InlineKeyboardMarkup, error) {
	// Обрабатываем общую логику команды start
	_, question, err := h.ProcessStartCommand(telegramID, title)
	if err != nil {
		return nil, nil, err
	}

	// Создаем клавиатуру
	keyboard := h.CreateQuestionKeyboard()

	return question, keyboard, nil
}

// SendQuestion отправляет вопрос (с картинкой или без)
func (h *BaseHandler) SendQuestion(chatID int64, question *models.Question, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	// Проверяем наличие картинки вопроса
	if question.QuestionPicture != nil && question.QuestionPicture.Path != nil {
		// Формируем URL картинки
		photoURL, err := h.GetPictureURL(*question.QuestionPicture.Path)
		if err != nil {
			fmt.Printf("Failed to get picture URL: %v (path: %s)\n", err, *question.QuestionPicture.Path)
			// Если не удалось получить картинку, отправляем текстовое сообщение
			questionText := h.FormatQuestionText(question)
			return h.SendMessage(chatID, questionText, keyboard)
		}

		// Формируем текст вопроса
		questionText := h.FormatQuestionText(question)

		// Отправляем фото с подписью
		return h.SendPhoto(chatID, photoURL, questionText, keyboard)
	} else {
		// Формируем текст вопроса
		questionText := h.FormatQuestionText(question)

		// Отправляем текстовое сообщение
		return h.SendMessage(chatID, questionText, keyboard)
	}
}

// ProcessUserReaction обрабатывает реакцию пользователя на вопрос
func (h *BaseHandler) ProcessUserReaction(chatID uint, questionID uint, reactionType string) error {
	// Создаем реакцию
	err := h.reactionRepo.CreateOrUpdateReaction(chatID, questionID, reactionType)
	if err != nil {
		return fmt.Errorf("failed to create reaction: %v", err)
	}

	// Уменьшаем баланс
	err = h.DecreaseBalance(chatID)
	if err != nil {
		return fmt.Errorf("failed to decrease balance: %v", err)
	}

	// Очищаем ожидание ответа
	err = h.ClearWaitingAnswer(chatID)
	if err != nil {
		fmt.Printf("Failed to clear waiting answer: %v (chat_id: %d)\n", err, chatID)
	}

	// Обновляем рейтинг вопроса после любой реакции
	err = h.questionRepo.UpdateQuestionRating(questionID)
	if err != nil {
		fmt.Printf("Failed to update question rating: %v (question_id: %d)\n", err, questionID)
		// Не возвращаем ошибку, так как основная логика уже выполнена
	}

	return nil
}

// ProcessSkipReaction обрабатывает реакцию "пропустить"
func (h *BaseHandler) ProcessSkipReaction(chatID uint, questionID uint) error {
	return h.ProcessUserReaction(chatID, questionID, "skip")
}

// ProcessFailReaction обрабатывает реакцию "показать ответ"
func (h *BaseHandler) ProcessFailReaction(chatID uint, questionID uint) error {
	return h.ProcessUserReaction(chatID, questionID, "fail")
}

// ProcessFinishReaction обрабатывает реакцию "закончить"
func (h *BaseHandler) ProcessFinishReaction(chatID uint, questionID uint) error {
	// Для finish используем тот же ProcessUserReaction с типом "fail"
	return h.ProcessUserReaction(chatID, questionID, "fail")
}

// SendMessage отправляет текстовое сообщение
func (h *BaseHandler) SendMessage(chatID int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "MarkdownV2"

	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	_, err := h.bot.Send(msg)
	return err
}

// SendPhoto отправляет фото с подписью
func (h *BaseHandler) SendPhoto(chatID int64, photoURL string, caption string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	photoConfig := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(photoURL))
	photoConfig.Caption = caption
	photoConfig.ParseMode = "MarkdownV2"

	if keyboard != nil {
		photoConfig.ReplyMarkup = keyboard
	}

	_, err := h.bot.Send(photoConfig)
	return err
}

// AnswerCallbackQuery отвечает на callback query
func (h *BaseHandler) AnswerCallbackQuery(callbackID string) error {
	callbackConfig := tgbotapi.NewCallback(callbackID, "")
	_, err := h.bot.Send(callbackConfig)
	return err
}
