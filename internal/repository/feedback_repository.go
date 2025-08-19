package repository

import (
	"gorm.io/gorm"
	"qweasley/internal/database"
	"qweasley/internal/models"
)

// FeedbackRepository репозиторий для работы с обратной связью
type FeedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository создает новый репозиторий обратной связи
func NewFeedbackRepository() *FeedbackRepository {
	return &FeedbackRepository{
		db: database.GetDB(),
	}
}

// Create создает новую запись обратной связи
func (r *FeedbackRepository) Create(feedback *models.Feedback) error {
	return r.db.Create(feedback).Error
}
