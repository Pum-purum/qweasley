package repository

import (
	"gorm.io/gorm"
	"qweasley/internal/database"
	"qweasley/internal/models"
)

// QuestionRepository репозиторий для работы с вопросами
type QuestionRepository struct {
	db *gorm.DB
}

// NewQuestionRepository создает новый репозиторий вопросов
func NewQuestionRepository() *QuestionRepository {
	return &QuestionRepository{
		db: database.GetDB(),
	}
}

// GetRandomPublished получает случайный опубликованный вопрос
func (r *QuestionRepository) GetRandomPublished() (*models.Question, error) {
	var question models.Question
	err := r.db.Where("is_published = ?", true).
		Preload("Author").
		Order("RANDOM()").
		First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}
