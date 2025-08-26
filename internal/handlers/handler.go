package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CommandHandler интерфейс для обработчиков команд
type CommandHandler interface {
	Handle(message *tgbotapi.Message) error
	GetCommand() string
}

// CallbackHandler интерфейс для обработчиков callback'ов
type CallbackHandler interface {
	Handle(callback *tgbotapi.CallbackQuery) error
	GetCallbackData() string
}

// Registry реестр всех обработчиков
type Registry struct {
	commandHandlers  map[string]CommandHandler
	CallbackHandlers map[string]CallbackHandler
	textHandler      TextHandler
}

// NewRegistry создает новый реестр обработчиков
func NewRegistry(bot *tgbotapi.BotAPI) *Registry {
	registry := &Registry{
		commandHandlers:  make(map[string]CommandHandler),
		CallbackHandlers: make(map[string]CallbackHandler),
		textHandler:      NewTextResponseHandler(bot),
	}

	// Создаем обработчики команд
	startHandler := NewStartHandler(bot)
	balanceHandler := NewBalanceHandler(bot)
	rulesHandler := NewRulesHandler(bot)
	feedbackHandler := NewFeedbackHandler(bot)
	proposalHandler := NewProposalHandler(bot)

	// Регистрируем обработчики команд
	registry.RegisterCommand(startHandler)
	registry.RegisterCommand(balanceHandler)
	registry.RegisterCommand(rulesHandler)
	registry.RegisterCommand(feedbackHandler)
	registry.RegisterCommand(proposalHandler)

	// Регистрируем обработчики callback'ов
	registry.RegisterCallback(NewSkipCallback(bot))
	registry.RegisterCallback(NewFailCallback(bot))
	registry.RegisterCallback(NewContinueCallback(startHandler, bot))
	registry.RegisterCallback(NewFinishCallback(bot))

	return registry
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
func (r *Registry) HandleCommand(command string, message *tgbotapi.Message) error {
	if handler, exists := r.commandHandlers[command]; exists {
		return handler.Handle(message)
	}
	return fmt.Errorf("команда не найдена: %s", command)
}

// HandleCallback обрабатывает callback
func (r *Registry) HandleCallback(callbackData string, callback *tgbotapi.CallbackQuery) error {
	if handler, exists := r.CallbackHandlers[callbackData]; exists {
		return handler.Handle(callback)
	}
	return fmt.Errorf("callback не найден: %s", callbackData)
}

// HandleTextMessage обрабатывает текстовое сообщение
func (r *Registry) HandleTextMessage(message *tgbotapi.Message) error {
	return r.textHandler.Handle(message)
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
