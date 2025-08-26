package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// ProposalHandler обработчик команды /proposal
type ProposalHandler struct {
	*BaseHandler
}

// NewProposalHandler создает новый обработчик команды proposal
func NewProposalHandler(bot *tgbotapi.BotAPI) *ProposalHandler {
	return &ProposalHandler{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCommand возвращает название команды
func (h *ProposalHandler) GetCommand() string {
	return "proposal"
}

// Handle обрабатывает команду /proposal
func (h *ProposalHandler) Handle(message *tgbotapi.Message) error {
	// TODO: Реализовать форму предложения вопроса
	text := "Предложите свой вопрос для квиза\\! Формат:\n\n*Вопрос:* Ваш вопрос\n*Ответ:* Правильный ответ\n*Комментарий:* Дополнительная информация \\(необязательно\\)"
	return h.SendMessage(message.Chat.ID, text, nil)
}
