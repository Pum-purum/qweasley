package repository

import (
	"gorm.io/gorm"
	"qweasley/internal/database"
	"qweasley/internal/models"
	"time"
)

// ReactionRepository репозиторий для работы с реакциями
type ReactionRepository struct {
	db *gorm.DB
}

// NewReactionRepository создает новый репозиторий реакций
func NewReactionRepository() *ReactionRepository {
	return &ReactionRepository{
		db: database.GetDB(),
	}
}

// GetReactedQuestionIDs получает ID вопросов, на которые пользователь уже реагировал
func (r *ReactionRepository) GetReactedQuestionIDs(chatID uint) ([]uint, error) {
	var reactions []models.Reaction
	err := r.db.Table("reactions").Where("chat_id = ?", chatID).Find(&reactions).Error

	if err != nil {
		return nil, err
	}

	var questionIDs []uint
	for _, reaction := range reactions {
		questionIDs = append(questionIDs, reaction.QuestionID)
	}
	return questionIDs, err
}

// GetNotSkippedQuestionIDs получает ID вопросов, которые пользователь не пропускал
func (r *ReactionRepository) GetNotSkippedQuestionIDs(chatID uint) ([]uint, error) {
	var reactions []models.Reaction
	err := r.db.Table("reactions").Where("chat_id = ? AND (responsed_at IS NOT NULL OR failed_at IS NOT NULL)", chatID).Find(&reactions).Error
	if err != nil {
		return nil, err
	}

	var questionIDs []uint
	for _, reaction := range reactions {
		questionIDs = append(questionIDs, reaction.QuestionID)
	}
	return questionIDs, err
}

// CreateOrUpdateReaction создает или обновляет реакцию пользователя на вопрос
func (r *ReactionRepository) CreateOrUpdateReaction(chatID, questionID uint, reactionType string) error {
	now := time.Now()

	var reaction models.Reaction
	err := r.db.Where("chat_id = ? AND question_id = ?", chatID, questionID).First(&reaction).Error

	if err == gorm.ErrRecordNotFound {
		// Создаем новую реакцию
		reaction = models.Reaction{
			ChatID:     chatID,
			QuestionID: questionID,
		}
	}

	// Обновляем соответствующие поля в зависимости от типа реакции
	switch reactionType {
	case "skip":
		reaction.SkippedAt = &now
	case "response":
		reaction.ResponsedAt = &now
	case "fail":
		reaction.FailedAt = &now
	}

	return r.db.Save(&reaction).Error
}

// GetReaction получает реакцию пользователя на конкретный вопрос
func (r *ReactionRepository) GetReaction(chatID, questionID uint) (*models.Reaction, error) {
	var reaction models.Reaction
	err := r.db.Where("chat_id = ? AND question_id = ?", chatID, questionID).First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}
