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

// FindByID находит обратную связь по ID
func (r *FeedbackRepository) FindByID(id uint) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.db.Preload("Chat").First(&feedback, id).Error
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

// GetByChatID получает все записи обратной связи для чата
func (r *FeedbackRepository) GetByChatID(chatID uint) ([]models.Feedback, error) {
	var feedbacks []models.Feedback
	err := r.db.Where("chat_id = ?", chatID).Preload("Chat").Find(&feedbacks).Error
	return feedbacks, err
}

// GetAll получает все записи обратной связи
func (r *FeedbackRepository) GetAll() ([]models.Feedback, error) {
	var feedbacks []models.Feedback
	err := r.db.Preload("Chat").Find(&feedbacks).Error
	return feedbacks, err
}

// Save сохраняет изменения в обратной связи
func (r *FeedbackRepository) Save(feedback *models.Feedback) error {
	return r.db.Save(feedback).Error
}

// AddResponse добавляет ответ на обратную связь
func (r *FeedbackRepository) AddResponse(feedbackID uint, response string) error {
	return r.db.Model(&models.Feedback{}).Where("id = ?", feedbackID).Update("response", response).Error
}
