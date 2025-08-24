package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"qweasley/internal/models"
	"qweasley/internal/repository"
	"strings"
)

// StartHandler обработчик команды /start
type StartHandler struct {
	chatRepo     *repository.ChatRepository
	questionRepo *repository.QuestionRepository
	reactionRepo *repository.ReactionRepository
}

// NewStartHandler создает новый обработчик команды start
func NewStartHandler() *StartHandler {
	return &StartHandler{
		chatRepo:     repository.NewChatRepository(),
		questionRepo: repository.NewQuestionRepository(),
		reactionRepo: repository.NewReactionRepository(),
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
		fmt.Printf("Failed to get or create chat: %v (chat_id: %d, title: %s)\n", err, message.Chat.ID, message.Chat.Title)
		return "Произошла ошибка при обработке команды", nil
	}

	// Проверяем баланс
	if chat.Balance <= 0 {
		return "У вас закончились монеты\\. Пополните баланс командой /balance и ждем вас снова\\!", nil
	}

	// Получаем вопрос для пользователя
	question, err := h.questionRepo.GetQuestion(chat, h.reactionRepo)
	if err != nil {
		fmt.Printf("Failed to get question: %v (chat_id: %d, balance: %d)\n", err, chat.ID, chat.Balance)
		return "К сожалению, не удалось получить вопрос\\. Попробуйте позже\\!", nil
	}

	// Если вопросов больше нет
	if question == nil {
		return "Уоу, вы ответили на все вопросы\\! Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!", nil
	}

	// Формируем текст вопроса
	questionText := h.formatQuestionText(question)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", fmt.Sprintf("skip:%d", question.ID)),
			tgbotapi.NewInlineKeyboardButtonData("Показать ответ", fmt.Sprintf("fail:%d", question.ID)),
			tgbotapi.NewInlineKeyboardButtonData("Закончить", fmt.Sprintf("finish:%d", question.ID)),
		),
	)

	return questionText, &keyboard
}

// HandleWithPhoto обрабатывает команду /start с отправкой фото
func (h *StartHandler) HandleWithPhoto(message *tgbotapi.Message) (*tgbotapi.PhotoConfig, error) {
	// Получаем или создаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat: %v (chat_id: %d, title: %s)\n", err, message.Chat.ID, message.Chat.Title)
		return nil, err
	}

	// Проверяем баланс
	if chat.Balance <= 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Получаем вопрос для пользователя
	question, err := h.questionRepo.GetQuestion(chat, h.reactionRepo)
	if err != nil {
		fmt.Printf("Failed to get question: %v (chat_id: %d, balance: %d)\n", err, chat.ID, chat.Balance)
		return nil, err
	}

	// Если вопросов больше нет
	if question == nil {
		return nil, fmt.Errorf("no questions available")
	}

	// Проверяем наличие картинки вопроса
	if question.QuestionPicture != nil && question.QuestionPicture.Path != nil {
		// Формируем URL картинки
		photoURL, err := h.getPictureURL(*question.QuestionPicture.Path)
		if err != nil {
			fmt.Printf("Failed to get picture URL: %v (path: %s)\n", err, *question.QuestionPicture.Path)
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
				tgbotapi.NewInlineKeyboardButtonData("Пропустить", fmt.Sprintf("skip:%d", question.ID)),
				tgbotapi.NewInlineKeyboardButtonData("Показать ответ", fmt.Sprintf("fail:%d", question.ID)),
				tgbotapi.NewInlineKeyboardButtonData("Закончить", fmt.Sprintf("finish:%d", question.ID)),
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

// HandleTextResponse обрабатывает текстовый ответ на вопрос
func (h *StartHandler) HandleTextResponse(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем или создаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat for text response: %v (chat_id: %d)\n", err, message.Chat.ID)
		return "Произошла ошибка при обработке ответа", nil
	}

	// Получаем активный вопрос для пользователя (без реакции)
	question, err := h.questionRepo.GetQuestion(chat, h.reactionRepo)
	if err != nil {
		fmt.Printf("Failed to get question for text answer: %v (chat_id: %d, user_answer: %s)\n", err, message.Chat.ID, message.Text)
		return "Произошла ошибка при проверке ответа", nil
	}

	// Если вопросов больше нет
	if question == nil {
		return "Уоу, вы ответили на все вопросы\\! Приходите завтра\\!", nil
	}

	// Проверяем ответ
	userAnswer := strings.ToLower(strings.TrimSpace(message.Text))
	correctAnswer := strings.ToLower(strings.TrimSpace(question.Answer))

	if userAnswer == correctAnswer {
		// Создаем реакцию "response"
		err = h.reactionRepo.CreateOrUpdateReaction(chat.ID, question.ID, "response")
		if err != nil {
			fmt.Printf("Failed to create response reaction: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, question.ID)
			return "Произошла ошибка при обработке ответа", nil
		}

		// Уменьшаем баланс
		err = h.chatRepo.DecreaseBalance(chat.ID)
		if err != nil {
			fmt.Printf("Failed to decrease balance for text answer: %v (chat_id: %d, question_id: %d)\n", err, chat.ID, question.ID)
			return "Произошла ошибка при обработке ответа", nil
		}

		// Формируем ответ
		responseText := "*Это правильный ответ\\!*"
		if question.Comment != nil {
			responseText += "\n\n" + h.escapeMarkdown(*question.Comment)
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Продолжаем", "continue"),
				tgbotapi.NewInlineKeyboardButtonData("Закончить", fmt.Sprintf("finish:%d", question.ID)),
			),
		)
		return responseText, &keyboard
	}

	return "Ответ неверный\\. Попробуйте еще раз", nil
}
