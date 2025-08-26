package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"qweasley/internal/models"
	"qweasley/internal/repository"
)

// FeedbackHandler обработчик команды /feedback
type FeedbackHandler struct {
	*BaseHandler
	feedbackRepo *repository.FeedbackRepository
}

// NewFeedbackHandler создает новый обработчик команды feedback
func NewFeedbackHandler(bot *tgbotapi.BotAPI) *FeedbackHandler {
	return &FeedbackHandler{
		BaseHandler:  NewBaseHandler(bot),
		feedbackRepo: repository.NewFeedbackRepository(),
	}
}

// GetCommand возвращает название команды
func (h *FeedbackHandler) GetCommand() string {
	return "feedback"
}

// Handle обрабатывает команду /feedback
func (h *FeedbackHandler) Handle(message *tgbotapi.Message) error {
	// Получаем или создаем чат пользователя
	chat, err := h.GetOrCreateChat(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat: %v\n", err)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	// Если это текстовое сообщение (не команда), сохраняем его как обратную связь
	if !message.IsCommand() && message.Text != "" && len(message.Text) >= 3 {
		feedback := &models.Feedback{
			Text:   message.Text,
			ChatID: chat.ID,
		}

		err = h.feedbackRepo.Create(feedback)
		if err != nil {
			fmt.Printf("Failed to create feedback: %v\n", err)
			return h.SendMessage(message.Chat.ID, "Произошла ошибка при сохранении сообщения", nil)
		}

		return h.SendMessage(message.Chat.ID, "Ваше сообщение принято\\! Спасибо\\!", nil)
	}

	text := "Напишите ваше сообщение администрации\\. Мы обязательно его прочитаем и ответим\\!"
	return h.SendMessage(message.Chat.ID, text, nil)
}
