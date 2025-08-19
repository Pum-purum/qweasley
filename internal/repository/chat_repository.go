package repository

import (
	"gorm.io/gorm"
	"qweasley/internal/database"
	"qweasley/internal/models"
)

// ChatRepository репозиторий для работы с чатами
type ChatRepository struct {
	db *gorm.DB
}

// NewChatRepository создает новый репозиторий чатов
func NewChatRepository() *ChatRepository {
	return &ChatRepository{
		db: database.GetDB(),
	}
}

// FindByTelegramID находит чат по Telegram ID
func (r *ChatRepository) FindByTelegramID(telegramID int64) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Where("telegram_id = ?", telegramID).First(&chat).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// Create создает новый чат
func (r *ChatRepository) Create(chat *models.Chat) error {
	return r.db.Create(chat).Error
}

// Save сохраняет изменения в чате
func (r *ChatRepository) Save(chat *models.Chat) error {
	return r.db.Save(chat).Error
}

// GetOrCreate получает существующий чат или создает новый
func (r *ChatRepository) GetOrCreate(telegramID int64, title *string) (*models.Chat, error) {
	chat, err := r.FindByTelegramID(telegramID)
	if err == nil {
		// Чат найден, обновляем заголовок если нужно
		if title != nil && (chat.Title == nil || *chat.Title != *title) {
			chat.Title = title
			err = r.Save(chat)
		}
		return chat, err
	}

	// Чат не найден, создаем новый
	chat = &models.Chat{
		TelegramID: telegramID,
		Title:      title,
		Balance:    30, // Начальный баланс
	}

	err = r.Create(chat)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

// UpdateBalance обновляет баланс чата
func (r *ChatRepository) UpdateBalance(chatID uint, balance int) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", chatID).Update("balance", balance).Error
}

// DecreaseBalance уменьшает баланс на 1
func (r *ChatRepository) DecreaseBalance(chatID uint) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", chatID).Update("balance", gorm.Expr("balance - 1")).Error
}
