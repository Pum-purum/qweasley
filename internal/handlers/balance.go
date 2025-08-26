package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BalanceHandler обработчик команды /balance
type BalanceHandler struct {
	*BaseHandler
}

// NewBalanceHandler создает новый обработчик команды balance
func NewBalanceHandler(bot *tgbotapi.BotAPI) *BalanceHandler {
	return &BalanceHandler{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCommand возвращает название команды
func (h *BalanceHandler) GetCommand() string {
	return "balance"
}

// Handle обрабатывает команду /balance
func (h *BalanceHandler) Handle(message *tgbotapi.Message) error {
	// Получаем или создаем чат пользователя
	chat, err := h.GetOrCreateChat(message.Chat.ID, &message.Chat.Title)
	if err != nil {
		fmt.Printf("Failed to get or create chat: %v\n", err)
		return h.SendMessage(message.Chat.ID, "Произошла ошибка при получении баланса", nil)
	}

	text := fmt.Sprintf("*Ваш баланс: %d монет\\.*\n\nПополнить баланс вы можете, предложив свой вопрос через соответствующую команду меню\\. В случае, если вопрос пройдет модерацию, он будет опубликован в боте и ваш счет будет пополнен на 10 монет\\. Если вы готовы приобрести монеты за деньги по курсу 1 монета \\= 10 рублей, свяжитесь с администрацией через команду \\/feedback", chat.Balance)

	return h.SendMessage(message.Chat.ID, text, nil)
}
