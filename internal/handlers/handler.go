package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// CommandHandler интерфейс для всех обработчиков команд
type CommandHandler interface {
	Handle(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup)
	GetCommand() string
}

// CallbackHandler интерфейс для обработчиков callback'ов
type CallbackHandler interface {
	Handle(callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup)
	GetCallbackData() string
}

// Registry реестр всех обработчиков
type Registry struct {
	commandHandlers  map[string]CommandHandler
	callbackHandlers map[string]CallbackHandler
}

// NewRegistry создает новый реестр обработчиков
func NewRegistry() *Registry {
	return &Registry{
		commandHandlers:  make(map[string]CommandHandler),
		callbackHandlers: make(map[string]CallbackHandler),
	}
}

// RegisterCommand регистрирует обработчик команды
func (r *Registry) RegisterCommand(handler CommandHandler) {
	r.commandHandlers[handler.GetCommand()] = handler
}

// RegisterCallback регистрирует обработчик callback'а
func (r *Registry) RegisterCallback(handler CallbackHandler) {
	r.callbackHandlers[handler.GetCallbackData()] = handler
}

// HandleCommand обрабатывает команду
func (r *Registry) HandleCommand(command string, message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	if handler, exists := r.commandHandlers[command]; exists {
		return handler.Handle(message)
	}
	return "Неизвестная команда. Доступные команды: /start, /balance, /rules, /feedback, /proposal", nil
}

// HandleCallback обрабатывает callback
func (r *Registry) HandleCallback(callbackData string, callback *tgbotapi.CallbackQuery) (string, *tgbotapi.InlineKeyboardMarkup) {
	if handler, exists := r.callbackHandlers[callbackData]; exists {
		return handler.Handle(callback)
	}
	return "Неизвестный callback", nil
}
