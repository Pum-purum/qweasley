package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// StartHandler обработчик команды /start
type StartHandler struct{}

// NewStartHandler создает новый обработчик команды start
func NewStartHandler() *StartHandler {
	return &StartHandler{}
}

// GetCommand возвращает название команды
func (h *StartHandler) GetCommand() string {
	return "start"
}

// Handle обрабатывает команду /start
func (h *StartHandler) Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем ID пользователя из сообщения
	userID := message.From.ID

	// TODO: Проверить баланс пользователя (userID)
	// TODO: Получить случайный вопрос из базы
	// TODO: Создать пользователя если не существует (30 монет)

	// Временно используем userID в логе (можно убрать позже)
	_ = userID

	// Заглушка - показываем пример вопроса
	questionText := "*Вопрос:*\n\nКакая планета ближайшая к Солнцу?"

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пропустить", "skip"),
			tgbotapi.NewInlineKeyboardButtonData("Показать ответ", "fail"),
			tgbotapi.NewInlineKeyboardButtonData("Закончить", "finish"),
		),
	)

	return questionText, &keyboard
}
