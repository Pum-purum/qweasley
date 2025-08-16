package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BalanceHandler обработчик команды /balance
type BalanceHandler struct{}

// NewBalanceHandler создает новый обработчик команды balance
func NewBalanceHandler() *BalanceHandler {
	return &BalanceHandler{}
}

// GetCommand возвращает название команды
func (h *BalanceHandler) GetCommand() string {
	return "balance"
}

// Handle обрабатывает команду /balance
func (h *BalanceHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Получить баланс из базы данных
	balance := 30 // Заглушка

	text := fmt.Sprintf("*Ваш баланс: %d монет\\.*\n\nПополнить баланс вы можете, предложив свой вопрос через соответствующую команду меню\\. В случае, если вопрос пройдет модерацию, он будет опубликован в боте и ваш счет будет пополнен на 10 монет\\. Если вы готовы приобрести монеты за деньги по курсу 1 монета \\= 10 рублей, свяжитесь с администрацией через команду \\/feedback", balance)

	return text, nil
}
