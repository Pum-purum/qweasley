package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// SkipCallback обработчик callback'а "skip"
type SkipCallback struct {
	startHandler *StartHandler
}

// NewSkipCallback создает новый обработчик callback'а skip
func NewSkipCallback(startHandler *StartHandler) *SkipCallback {
	return &SkipCallback{startHandler: startHandler}
}

// GetCallbackData возвращает данные callback'а
func (h *SkipCallback) GetCallbackData() string {
	return "skip"
}

// Handle обрабатывает callback "skip"
func (h *SkipCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Пропустить вопрос, списать монету, показать следующий
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}
	return h.startHandler.Handle(message)
}

// FailCallback обработчик callback'а "fail"
type FailCallback struct{}

// NewFailCallback создает новый обработчик callback'а fail
func NewFailCallback() *FailCallback {
	return &FailCallback{}
}

// GetCallbackData возвращает данные callback'а
func (h *FailCallback) GetCallbackData() string {
	return "fail"
}

// Handle обрабатывает callback "fail"
func (h *FailCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Показать правильный ответ, списать монету
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Точно!", "continue"),
				tgbotapi.NewInlineKeyboardButtonData("Ладно, хватит", "finish"),
			},
		},
	}
	text := "*Правильный ответ:*\nМеркурий\n\nМеркурий \\- самая близкая к Солнцу планета Солнечной системы\\."
	return text, keyboard
}

// ContinueCallback обработчик callback'а "continue"
type ContinueCallback struct {
	startHandler *StartHandler
}

// NewContinueCallback создает новый обработчик callback'а continue
func NewContinueCallback(startHandler *StartHandler) *ContinueCallback {
	return &ContinueCallback{startHandler: startHandler}
}

// GetCallbackData возвращает данные callback'а
func (h *ContinueCallback) GetCallbackData() string {
	return "continue"
}

// Handle обрабатывает callback "continue"
func (h *ContinueCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	// TODO: Показать следующий вопрос
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: callback.Message.Chat,
	}
	return h.startHandler.Handle(message)
}

// FinishCallback обработчик callback'а "finish"
type FinishCallback struct{}

// NewFinishCallback создает новый обработчик callback'а finish
func NewFinishCallback() *FinishCallback {
	return &FinishCallback{}
}

// GetCallbackData возвращает данные callback'а
func (h *FinishCallback) GetCallbackData() string {
	return "finish"
}

// Handle обрабатывает callback "finish"
func (h *FinishCallback) Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	text := "Приходите завтра\\! Новые интересные вопросы появляются каждый день\\!"
	return text, nil
}
