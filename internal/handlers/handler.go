package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

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
	CallbackHandlers map[string]CallbackHandler
}

// NewRegistry создает новый реестр обработчиков
func NewRegistry() *Registry {
	return &Registry{
		commandHandlers:  make(map[string]CommandHandler),
		CallbackHandlers: make(map[string]CallbackHandler),
	}
}

// RegisterCommand регистрирует обработчик команды
func (r *Registry) RegisterCommand(handler CommandHandler) {
	r.commandHandlers[handler.GetCommand()] = handler
}

// RegisterCallback регистрирует обработчик callback'а
func (r *Registry) RegisterCallback(handler CallbackHandler) {
	r.CallbackHandlers[handler.GetCallbackData()] = handler
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
	// Парсим callback данные (формат: "action:questionID")
	parts := strings.Split(callbackData, ":")
	action := parts[0]

	if handler, exists := r.CallbackHandlers[action]; exists {
		return handler.Handle(callback)
	}
	return "Неизвестный callback", nil
}

// GetStartHandler возвращает обработчик команды start
func (r *Registry) GetStartHandler() *StartHandler {
	if handler, exists := r.commandHandlers["start"]; exists {
		if startHandler, ok := handler.(*StartHandler); ok {
			return startHandler
		}
	}
	return nil
}

// GetFailHandler возвращает обработчик callback'а fail
func (r *Registry) GetFailHandler() *FailCallback {
	if handler, exists := r.CallbackHandlers["fail"]; exists {
		if failHandler, ok := handler.(*FailCallback); ok {
			return failHandler
		}
	}
	return nil
}

// HandleTextMessage обрабатывает текстовое сообщение
func (r *Registry) HandleTextMessage(message *tgbotapi.Message) (string, *tgbotapi.InlineKeyboardMarkup) {
	// Получаем обработчик start для обработки текстовых ответов
	startHandler := r.GetStartHandler()
	if startHandler != nil {
		return startHandler.HandleTextResponse(message)
	}
	return "Начните квиз командой /start", nil
}
