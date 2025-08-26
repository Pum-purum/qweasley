package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// RulesHandler обработчик команды /rules
type RulesHandler struct {
	*BaseHandler
}

// NewRulesHandler создает новый обработчик команды rules
func NewRulesHandler(bot *tgbotapi.BotAPI) *RulesHandler {
	return &RulesHandler{
		BaseHandler: NewBaseHandler(bot),
	}
}

// GetCommand возвращает название команды
func (h *RulesHandler) GetCommand() string {
	return "rules"
}

// Handle обрабатывает команду /rules
func (h *RulesHandler) Handle(message *tgbotapi.Message) error {
	text := "*Правила*\n\n1\\. При первом контакте с ботом на ваш счет закидывается 30 монет\\.\n2\\. За каждый верно отвеченный вопрос со счета снимается 1 монета\\.\n3\\. Ответом является одно слово на русском языке в именительном падеже единственного числа, если в вопросе не указано иное\\.\n4\\. Если ответом является калька с иностранного языка, имеющая несколько вариантов написания, то правильным будет тот, который указан в Википедии\\.\n5\\. Регистр букв в ответе не имеет значения\\.\n6\\. За каждое нажатие кнопки Показать ответ со счета снимается 1 монета\\.\n7\\. Счет привязан не к пользователю, а к чату\\.\n8\\. Монеты со счета нельзя вернуть\\, но можно отдать другому чату\\, для этого напишите в форму обратной связи\\.\n9\\. Бот поставляется \"как есть\"\\. Администрация не несет ответственности за любые негативные последствия, прямо или косвенно вызванные использованием бота\\."

	return h.SendMessage(message.Chat.ID, text, nil)
}
