package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// ProposalHandler обработчик команды /proposal
type ProposalHandler struct{}

// NewProposalHandler создает новый обработчик команды proposal
func NewProposalHandler() *ProposalHandler {
	return &ProposalHandler{}
}

// GetCommand возвращает название команды
func (h *ProposalHandler) GetCommand() string {
	return "proposal"
}

// Handle обрабатывает команду /proposal
func (h *ProposalHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Реализовать форму предложения вопроса
	text := "Предложите свой вопрос для квиза\\! Формат:\n\n*Вопрос:* Ваш вопрос\n*Ответ:* Правильный ответ\n*Комментарий:* Дополнительная информация \\(необязательно\\)"
	return text, nil
}
