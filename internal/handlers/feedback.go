package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"qweasley/internal/models"
	"qweasley/internal/repository"
)

// FeedbackHandler обработчик команды /feedback
type FeedbackHandler struct {
	chatRepo     *repository.ChatRepository
	feedbackRepo *repository.FeedbackRepository
}

// NewFeedbackHandler создает новый обработчик команды feedback
func NewFeedbackHandler() *FeedbackHandler {
	return &FeedbackHandler{
		chatRepo:     repository.NewChatRepository(),
		feedbackRepo: repository.NewFeedbackRepository(),
	}
}

// GetCommand возвращает название команды
func (h *FeedbackHandler) GetCommand() string {
	return "feedback"
}

// Handle обрабатывает команду /feedback
func (h *FeedbackHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем или создаем чат пользователя
	chat, err := h.chatRepo.GetOrCreate(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		log.Printf("Failed to get or create chat: %v", err)
		return "Произошла ошибка при обработке команды", nil
	}

	// Если это текстовое сообщение (не команда), сохраняем его как обратную связь
	if !message.IsCommand() && message.Text != "" && len(message.Text) >= 3 {
		feedback := &models.Feedback{
			Text:   message.Text,
			ChatID: chat.ID,
		}

		err = h.feedbackRepo.Create(feedback)
		if err != nil {
			log.Printf("Failed to create feedback: %v", err)
			return "Произошла ошибка при сохранении сообщения", nil
		}

		return "Ваше сообщение принято\\! Спасибо\\!", nil
	}

	text := "Напишите ваше сообщение администрации\\. Мы обязательно его прочитаем и ответим\\!"
	return text, nil
}
