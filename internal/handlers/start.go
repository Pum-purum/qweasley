package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"qweasley/internal/models"
	"qweasley/internal/repository"
	"qweasley/internal/utils"
	"strings"
)

// StartHandler обработчик команды /start
type StartHandler struct {
	chatRepo     *repository.ChatRepository
	questionRepo *repository.QuestionRepository
	reactionRepo *repository.ReactionRepository
	sessionMgr   *SessionManager
}

// NewStartHandler создает новый обработчик команды start
func NewStartHandler(sessionMgr *SessionManager) *StartHandler {
	return &StartHandler{
		chatRepo:     repository.NewChatRepository(),
		questionRepo: repository.NewQuestionRepository(),
		reactionRepo: repository.NewReactionRepository(),
		sessionMgr:   sessionMgr,
	}
}

// GetCommand возвращает название команды
func (h *StartHandler) GetCommand() string {
	return "start"
}

// Handle обрабатывает команду /start
func (h *StartHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем или создаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get or create chat", map[string]interface{}{
			"chat_id": message.Chat.ID,
			"title":   message.Chat.Title,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Проверяем баланс
	if chat.Balance <= 0 {
		return "У вас закончились монеты\\. Пополните баланс командой /balance и ждем вас снова\\!", nil
	}

	// Получаем вопрос для пользователя
	question, err := h.questionRepo.GetQuestion(chat, h.reactionRepo)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get question", map[string]interface{}{
			"chat_id": chat.ID,
			"balance": chat.Balance,
		})
		return "К сожалению, не удалось получить вопрос\\. Попробуйте позже\\!", nil
	}

	// Если вопросов больше нет
	if question == nil {
		return "Уоу, вы ответили на все вопросы\\! Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!", nil
	}

	// Сохраняем сессию с текущим вопросом
	h.sessionMgr.SetSession(message.Chat.ID, question.ID)

	// Формируем текст вопроса
	questionText := h.formatQuestionText(question)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip"),
			tgbotapi.NewInlineKeyboardButtonData("Показать ответ", "fail"),
			tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
		),
	)

	return questionText, &keyboard
}

// HandleWithPhoto обрабатывает команду /start с отправкой фото
func (h *StartHandler) HandleWithPhoto(message *tgbotapi.Message) (*tgbotapi.PhotoConfig, error) {
	// Получаем или создаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get or create chat", map[string]interface{}{
			"chat_id": message.Chat.ID,
			"title":   message.Chat.Title,
		})
		return nil, err
	}

	// Проверяем баланс
	if chat.Balance <= 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Получаем вопрос для пользователя
	question, err := h.questionRepo.GetQuestion(chat, h.reactionRepo)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get question", map[string]interface{}{
			"chat_id": chat.ID,
			"balance": chat.Balance,
		})
		return nil, err
	}

	// Если вопросов больше нет
	if question == nil {
		return nil, fmt.Errorf("no questions available")
	}

	// Сохраняем сессию с текущим вопросом
	h.sessionMgr.SetSession(message.Chat.ID, question.ID)

	// Проверяем наличие картинки вопроса
	if question.QuestionPicture != nil && question.QuestionPicture.Path != nil {
		// Формируем URL картинки
		photoURL, err := h.getPictureURL(*question.QuestionPicture.Path)
		if err != nil {
			utils.LogErrorWithContext(err, "Failed to get picture URL", map[string]interface{}{
				"path": *question.QuestionPicture.Path,
			})
			return nil, err
		}

		// Формируем текст вопроса
		caption := h.formatQuestionText(question)

		// Создаем конфигурацию фото
		photoConfig := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileURL(photoURL))
		photoConfig.Caption = caption
		photoConfig.ParseMode = "MarkdownV2"

		// Добавляем клавиатуру
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip"),
				tgbotapi.NewInlineKeyboardButtonData("Показать ответ", "fail"),
				tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
			),
		)
		photoConfig.ReplyMarkup = keyboard

		return &photoConfig, nil
	}

	return nil, fmt.Errorf("no photo available")
}

// getPictureURL формирует URL картинки
func (h *StartHandler) getPictureURL(path string) (string, error) {
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

// formatQuestionText форматирует текст вопроса для отправки
func (h *StartHandler) formatQuestionText(question *models.Question) string {
	// Экранируем специальные символы для Markdown
	text := h.escapeMarkdown(question.Text)

	// Добавляем рейтинг, если он есть
	if question.Rating != nil {
		add := fmt.Sprintf("На этот вопрос отвечают %d%% пользователей", *question.Rating)
		text += "\n\n_" + h.escapeMarkdown(add) + "_"
	}

	return text
}

// escapeMarkdown экранирует специальные символы для Markdown
func (h *StartHandler) escapeMarkdown(text string) string {
	specialChars := []string{"?", "!", "_", "*", "[", "]", "(", ")", "~", "`", ">", "<", "&", "#", "+", "-", "=", "|", "{", "}", "."}

	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}

	return text
}
