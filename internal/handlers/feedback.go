package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"qweasley/internal/models"
	"qweasley/internal/repository"
	"time"
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

	// Устанавливаем состояние ожидания обратной связи (30 минут)
	err = h.SetWaitingFeedback(chat.ID, 30*time.Minute)
	if err != nil {
		fmt.Printf("Failed to set waiting feedback: %v\n", err)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке команды", nil)
	}

	text := "Если у вас есть вопросы, предложения или жалобы, напишите их следующим сообщением\\. Мы обязательно их увидим\\."
	return h.SendMessage(message.Chat.ID, text, nil)
}

// HandleFeedbackMessage обрабатывает сообщение обратной связи
func (h *FeedbackHandler) HandleFeedbackMessage(message *tgbotapi.Message) error {
	// Получаем или создаем чат пользователя
	chat, err := h.GetOrCreateChat(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat: %v\n", err)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при обработке сообщения", nil)
	}

	// Проверяем, находится ли чат в состоянии ожидания обратной связи
	if !chat.IsWaitingFeedback() {
		return nil // Не в состоянии обратной связи, игнорируем
	}

	// Проверяем валидность сообщения
	if message.Text == "" || len(message.Text) < 3 {
		h.ClearWaitingFeedback(chat.ID)
		return h.SendMessage(message.Chat.ID, "К сожалению, это некорректное сообщение\\.", nil)
	}

	// Создаем обратную связь
	feedback := &models.Feedback{
		Text:   message.Text,
		ChatID: chat.ID,
	}

	err = h.feedbackRepo.Create(feedback)
	if err != nil {
		fmt.Printf("Failed to create feedback: %v\n", err)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при сохранении сообщения", nil)
	}

	// Очищаем состояние
	err = h.ClearWaitingFeedback(chat.ID)
	if err != nil {
		fmt.Printf("Failed to clear waiting feedback: %v\n", err)
	}

	// Отправляем подтверждение пользователю
	err = h.SendMessage(message.Chat.ID, "Ваше сообщение принято\\! Спасибо\\!", nil)
	if err != nil {
		return err
	}

	// Отправляем уведомление администратору
	adminChatID := os.Getenv("ADMIN_CHAT_ID")
	if adminChatID != "" {
		adminID := int64(0)
		fmt.Sscanf(adminChatID, "%d", &adminID)
		if adminID != 0 {
			h.SendMessage(adminID, "Новое сообщение в форме обратной связи", nil)
		}
	}

	return nil
}
