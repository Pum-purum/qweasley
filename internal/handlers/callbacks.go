package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"qweasley/internal/repository"
	"qweasley/internal/utils"
	"strings"
)

// SkipCallback обработчик callback'а "skip"
type SkipCallback struct {
	startHandler *StartHandler
	chatRepo     *repository.ChatRepository
	reactionRepo *repository.ReactionRepository
}

// NewSkipCallback создает новый обработчик callback'а skip
func NewSkipCallback(startHandler *StartHandler) *SkipCallback {
	return &SkipCallback{
		startHandler: startHandler,
		chatRepo:     repository.NewChatRepository(),
		reactionRepo: repository.NewReactionRepository(),
	}
}

// GetCallbackData возвращает данные callback'а
func (h *SkipCallback) GetCallbackData() string {
	return "skip"
}

// Handle обрабатывает callback "skip"
func (h *SkipCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем сессию пользователя
	session := h.startHandler.sessionMgr.GetSession(callback.Message.Chat.ID)
	if session == nil {
		return "Сессия истекла\\. Начните заново командой /start", nil
	}

	// Получаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get or create chat in skip callback", map[string]interface{}{
			"chat_id":             callback.Message.Chat.ID,
			"session_question_id": session.QuestionID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Создаем реакцию "пропустить"
	err = h.reactionRepo.CreateOrUpdateReaction(chat.ID, session.QuestionID, "skip")
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to create skip reaction", map[string]interface{}{
			"chat_id":     chat.ID,
			"question_id": session.QuestionID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Уменьшаем баланс
	err = h.chatRepo.DecreaseBalance(chat.ID)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to decrease balance in skip callback", map[string]interface{}{
			"chat_id":     chat.ID,
			"question_id": session.QuestionID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Показываем следующий вопрос
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}
	return h.startHandler.Handle(message)
}

// FailCallback обработчик callback'а "fail"
type FailCallback struct {
	startHandler *StartHandler
	chatRepo     *repository.ChatRepository
	reactionRepo *repository.ReactionRepository
	questionRepo *repository.QuestionRepository
}

// NewFailCallback создает новый обработчик callback'а fail
func NewFailCallback(startHandler *StartHandler) *FailCallback {
	return &FailCallback{
		startHandler: startHandler,
		chatRepo:     repository.NewChatRepository(),
		reactionRepo: repository.NewReactionRepository(),
		questionRepo: repository.NewQuestionRepository(),
	}
}

// GetCallbackData возвращает данные callback'а
func (h *FailCallback) GetCallbackData() string {
	return "fail"
}

// Handle обрабатывает callback "fail"
func (h *FailCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем сессию пользователя
	session := h.startHandler.sessionMgr.GetSession(callback.Message.Chat.ID)
	if session == nil {
		return "Сессия истекла\\. Начните заново командой /start", nil
	}

	// Получаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get or create chat in fail callback", map[string]interface{}{
			"chat_id":             callback.Message.Chat.ID,
			"session_question_id": session.QuestionID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Получаем вопрос
	question, err := h.questionRepo.GetByID(session.QuestionID)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get question in fail callback", map[string]interface{}{
			"question_id": session.QuestionID,
			"chat_id":     chat.ID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Создаем реакцию "fail"
	err = h.reactionRepo.CreateOrUpdateReaction(chat.ID, session.QuestionID, "fail")
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to create fail reaction", map[string]interface{}{
			"chat_id":     chat.ID,
			"question_id": session.QuestionID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Уменьшаем баланс
	err = h.chatRepo.DecreaseBalance(chat.ID)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to decrease balance in fail callback", map[string]interface{}{
			"chat_id":     chat.ID,
			"question_id": session.QuestionID,
		})
		return "Произошла ошибка при обработке команды", nil
	}

	// Формируем ответ с правильным ответом
	answerText := "*Правильный ответ:*\n" + h.escapeMarkdown(question.Answer)
	if question.Comment != nil {
		answerText += "\n\n" + h.escapeMarkdown(*question.Comment)
	}

	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Точно!", "continue"),
				tgbotapi.NewInlineKeyboardButtonData("Ладно, хватит", "finish"),
			},
		},
	}

	return answerText, keyboard
}

// HandleWithPhoto обрабатывает callback "fail" с отправкой фото
func (h *FailCallback) HandleWithPhoto(callback *tgbotapi.CallbackQuery) (*tgbotapi.PhotoConfig, error) {
	// Получаем сессию пользователя
	session := h.startHandler.sessionMgr.GetSession(callback.Message.Chat.ID)
	if session == nil {
		return nil, fmt.Errorf("session expired")
	}

	// Получаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(callback.Message.Chat.ID, &callback.Message.Chat.Title)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get or create chat in fail callback", map[string]interface{}{
			"chat_id":             callback.Message.Chat.ID,
			"session_question_id": session.QuestionID,
		})
		return nil, err
	}

	// Получаем вопрос
	question, err := h.questionRepo.GetByID(session.QuestionID)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to get question in fail callback", map[string]interface{}{
			"question_id": session.QuestionID,
			"chat_id":     chat.ID,
		})
		return nil, err
	}

	// Создаем реакцию "fail"
	err = h.reactionRepo.CreateOrUpdateReaction(chat.ID, session.QuestionID, "fail")
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to create fail reaction", map[string]interface{}{
			"chat_id":     chat.ID,
			"question_id": session.QuestionID,
		})
		return nil, err
	}

	// Уменьшаем баланс
	err = h.chatRepo.DecreaseBalance(chat.ID)
	if err != nil {
		utils.LogErrorWithContext(err, "Failed to decrease balance in fail callback", map[string]interface{}{
			"chat_id":     chat.ID,
			"question_id": session.QuestionID,
		})
		return nil, err
	}

	// Проверяем наличие картинки ответа
	if question.AnswerPicture != nil && question.AnswerPicture.Path != nil {
		// Формируем URL картинки
		photoURL, err := h.getPictureURL(*question.AnswerPicture.Path)
		if err != nil {
			utils.LogErrorWithContext(err, "Failed to get picture URL in fail callback", map[string]interface{}{
				"path": *question.AnswerPicture.Path,
			})
			return nil, err
		}

		// Формируем ответ с правильным ответом
		caption := "*Правильный ответ:*\n" + h.escapeMarkdown(question.Answer)
		if question.Comment != nil {
			caption += "\n\n" + h.escapeMarkdown(*question.Comment)
		}

		// Создаем конфигурацию фото
		photoConfig := tgbotapi.NewPhoto(callback.Message.Chat.ID, tgbotapi.FileURL(photoURL))
		photoConfig.Caption = caption
		photoConfig.ParseMode = "MarkdownV2"

		// Добавляем клавиатуру
		keyboard := &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("Точно!", "continue"),
					tgbotapi.NewInlineKeyboardButtonData("Ладно, хватит", "finish"),
				},
			},
		}
		photoConfig.ReplyMarkup = keyboard

		return &photoConfig, nil
	}

	return nil, fmt.Errorf("no photo available")
}

// getPictureURL формирует URL картинки
func (h *FailCallback) getPictureURL(path string) (string, error) {
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
func (h *FailCallback) escapeMarkdown(text string) string {
	specialChars := []string{"?", "!", "_", "*", "[", "]", "(", ")", "~", "`", ">", "<", "&", "#", "+", "-", "=", "|", "{", "}", "."}

	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}

	return text
}

// ContinueCallback обработчик callback'а "continue"
type ContinueCallback struct {
	startHandler *StartHandler
}

// NewContinueCallback создает новый обработчик callback'а continue
func NewContinueCallback(startHandler *StartHandler) *ContinueCallback {
	return &ContinueCallback{startHandler: startHandler}
}

// GetCallbackData возвращает данные callback'а
func (h *ContinueCallback) GetCallbackData() string {
	return "continue"
}

// Handle обрабатывает callback "continue"
func (h *ContinueCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Показать следующий вопрос
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}
	return h.startHandler.Handle(message)
}

// FinishCallback обработчик callback'а "finish"
type FinishCallback struct {
	startHandler *StartHandler
	chatRepo     *repository.ChatRepository
	reactionRepo *repository.ReactionRepository
}

// NewFinishCallback создает новый обработчик callback'а finish
func NewFinishCallback(startHandler *StartHandler) *FinishCallback {
	return &FinishCallback{
		startHandler: startHandler,
		chatRepo:     repository.NewChatRepository(),
		reactionRepo: repository.NewReactionRepository(),
	}
}

// GetCallbackData возвращает данные callback'а
func (h *FinishCallback) GetCallbackData() string {
	return "finish"
}

// Handle обрабатывает callback "finish"
func (h *FinishCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем сессию пользователя
	session := h.startHandler.sessionMgr.GetSession(callback.Message.Chat.ID)
	if session != nil {
		// Получаем чат пользователя
		chat, err := h.chatRepo.GetOrCreate(callback.Message.Chat.ID, &callback.Message.Chat.Title)
		if err == nil {
			// Создаем реакцию "fail" если еще не создана
			h.reactionRepo.CreateOrUpdateReaction(chat.ID, session.QuestionID, "fail")
			// Уменьшаем баланс
			h.chatRepo.DecreaseBalance(chat.ID)
		}
		// Очищаем сессию
		h.startHandler.sessionMgr.ClearSession(callback.Message.Chat.ID)
	}

	text := "Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!"
	return text, nil
}
