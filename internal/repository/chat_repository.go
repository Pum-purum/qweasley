package repository

import (
	"gorm.io/gorm"
	"qweasley/internal/database"
	"qweasley/internal/models"
	"time"
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

// GetOrCreate получает существующий чат или создает новый
func (r *ChatRepository) GetOrCreate(telegramID int64, title *string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Where("telegram_id = ?", telegramID).First(&chat).Error
	if err == nil {
		return &chat, err
	}

	// Чат не найден, создаем новый
	chat = models.Chat{
		TelegramID: telegramID,
		Title:      title,
		Balance:    30, // Начальный баланс
	}

	err = r.db.Create(&chat).Error
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

// SetWaitingAnswer устанавливает ожидание ответа на вопрос
func (r *ChatRepository) SetWaitingAnswer(chatID uint, questionID uint, expiresIn time.Duration) error {
	chat, err := r.GetByID(chatID)
	if err != nil {
		return err
	}

	chat.LastQuestionID = &questionID
	expiresAt := time.Now().UTC().Add(expiresIn)
	chat.ExpiresAt = &expiresAt

	return r.db.Save(chat).Error
}

// ClearWaitingAnswer очищает ожидание ответа
func (r *ChatRepository) ClearWaitingAnswer(chatID uint) error {
	chat, err := r.GetByID(chatID)
	if err != nil {
		return err
	}

	chat.LastQuestionID = nil
	chat.ExpiresAt = nil

	return r.db.Save(chat).Error
}

// UpdateBalance обновляет баланс чата
func (r *ChatRepository) UpdateBalance(chatID uint, balance int) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", chatID).Update("balance", balance).Error
}

// DecreaseBalance уменьшает баланс на 1
func (r *ChatRepository) DecreaseBalance(chatID uint) error {
	return r.db.Model(&models.Chat{}).Where("id = ?", chatID).Update("balance", gorm.Expr("balance - 1")).Error
}

// GetByID получает чат по ID
func (r *ChatRepository) GetByID(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.First(&chat, chatID).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// SetWaitingFeedback устанавливает ожидание обратной связи
func (r *ChatRepository) SetWaitingFeedback(chatID uint, expiresIn time.Duration) error {
	chat, err := r.GetByID(chatID)
	if err != nil {
		return err
	}

	expiresAt := time.Now().UTC().Add(expiresIn)
	chat.FeedbackExpiresAt = &expiresAt

	return r.db.Save(chat).Error
}

// ClearWaitingFeedback очищает ожидание обратной связи
func (r *ChatRepository) ClearWaitingFeedback(chatID uint) error {
	chat, err := r.GetByID(chatID)
	if err != nil {
		return err
	}

	chat.FeedbackExpiresAt = nil

	return r.db.Save(chat).Error
}
