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

// FindByID находит вопрос по ID
func (r *QuestionRepository) FindByID(id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.Preload("Author").Preload("QuestionPicture").Preload("AnswerPicture").First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// GetRandomPublished получает случайный опубликованный вопрос
func (r *QuestionRepository) GetRandomPublished() (*models.Question, error) {
	var question models.Question
	err := r.db.Where("is_published = ?", true).
		Preload("Author").
		Preload("QuestionPicture").
		Preload("AnswerPicture").
		Order("RANDOM()").
		First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// Create создает новый вопрос
func (r *QuestionRepository) Create(question *models.Question) error {
	return r.db.Create(question).Error
}

// Save сохраняет изменения в вопросе
func (r *QuestionRepository) Save(question *models.Question) error {
	return r.db.Save(question).Error
}

// GetUnpublishedQuestions получает неопубликованные вопросы
func (r *QuestionRepository) GetUnpublishedQuestions() ([]models.Question, error) {
	var questions []models.Question
	err := r.db.Where("is_published = ?", false).
		Preload("Author").
		Preload("QuestionPicture").
		Preload("AnswerPicture").
		Find(&questions).Error
	return questions, err
}

// PublishQuestion публикует вопрос
func (r *QuestionRepository) PublishQuestion(questionID uint) error {
	return r.db.Model(&models.Question{}).Where("id = ?", questionID).Update("is_published", true).Error
}

// GetQuestionsByAuthor получает вопросы по автору
func (r *QuestionRepository) GetQuestionsByAuthor(authorID uint) ([]models.Question, error) {
	var questions []models.Question
	err := r.db.Where("author_id = ?", authorID).
		Preload("QuestionPicture").
		Preload("AnswerPicture").
		Find(&questions).Error
	return questions, err
}
