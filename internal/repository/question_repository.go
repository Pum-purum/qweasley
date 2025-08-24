package repository

import (
	"fmt"
	"gorm.io/gorm"
	"math/rand"
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

// GetByID получает вопрос по ID
func (r *QuestionRepository) GetByID(id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.Preload("Author").
		Preload("QuestionPicture").
		Preload("AnswerPicture").
		Where("id = ?", id).
		First(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// GetQuestion получает вопрос для пользователя, исключая его собственные вопросы и уже отвеченные
func (r *QuestionRepository) GetQuestion(chat *models.Chat, reactionRepo *ReactionRepository) (*models.Question, error) {
	// Получаем ID вопросов, на которые пользователь уже реагировал
	reactedIDs, err := reactionRepo.GetReactedQuestionIDs(chat.ID)
	if err != nil {
		fmt.Printf("Failed to get reacted question IDs: %v (chat_id: %d)\n", err, chat.ID)
		return nil, err
	}

	// Строим запрос для получения вопросов
	query := r.db.Table("questions").Where("is_published = ?", true).
		Where("(author_id IS NULL OR author_id != ?)", chat.ID)

	// Исключаем уже отвеченные вопросы
	if len(reactedIDs) > 0 {
		query = query.Where("id NOT IN ?", reactedIDs)
	}

	// Получаем ID доступных вопросов
	var questionIDs []uint
	err = query.Pluck("id", &questionIDs).Error

	if err != nil {
		fmt.Printf("Failed to get question IDs: %v (chat_id: %d, reacted_count: %d)\n", err, chat.ID, len(reactedIDs))
		return nil, err
	}

	// Если нет новых вопросов, пробуем найти пропущенные
	if len(questionIDs) == 0 {
		notSkippedIDs, err := reactionRepo.GetNotSkippedQuestionIDs(chat.ID)
		if err != nil {
			return nil, err
		}

		query = r.db.Table("questions").Where("is_published = ?", true).
			Where("(author_id IS NULL OR author_id != ?)", chat.ID)

		if len(notSkippedIDs) > 0 {
			query = query.Where("id NOT IN ?", notSkippedIDs)
		}

		err = query.Pluck("id", &questionIDs).Error
		if err != nil {
			return nil, err
		}
	}

	// Если все еще нет вопросов, возвращаем nil
	if len(questionIDs) == 0 {
		return nil, nil
	}

	// Выбираем случайный вопрос из доступных
	randomIndex := rand.Intn(len(questionIDs))
	selectedID := questionIDs[randomIndex]

	// Получаем полную информацию о вопросе
	var question models.Question
	err = r.db.Preload("Author").
		Preload("QuestionPicture").
		Preload("AnswerPicture").
		Where("id = ?", selectedID).
		First(&question).Error
	if err != nil {
		return nil, err
	}

	return &question, nil
}
