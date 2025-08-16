package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// FeedbackHandler обработчик команды /feedback
type FeedbackHandler struct{}

// NewFeedbackHandler создает новый обработчик команды feedback
func NewFeedbackHandler() *FeedbackHandler {
	return &FeedbackHandler{}
}

// GetCommand возвращает название команды
func (h *FeedbackHandler) GetCommand() string {
	return "feedback"
}

// Handle обрабатывает команду /feedback
func (h *FeedbackHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Реализовать форму обратной связи
	text := "Напишите ваше сообщение администрации\\. Мы обязательно его прочитаем и ответим\\!"
	return text, nil
}
