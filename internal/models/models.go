package models

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Chat представляет чат пользователя
type Chat struct {
	ID         uint      `gorm:"primaryKey;column:id;default:nextval('chats_id_seq')" json:"id"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Balance    int       `gorm:"column:balance;default:0;not null" json:"balance"`
	TelegramID int64     `gorm:"column:telegram_id;uniqueIndex;not null" json:"telegram_id"`
	Title      *string   `gorm:"column:title" json:"title"`

	// Поля для отслеживания ожидания ответа
	LastQuestionID *uint      `gorm:"column:last_question_id" json:"last_question_id"`
	ExpiresAt      *time.Time `gorm:"column:expires_at" json:"expires_at"`
}

// TableName возвращает имя таблицы для Chat
func (Chat) TableName() string {
	return "chats"
}

// BeforeCreate хук перед созданием записи
func (c *Chat) BeforeCreate(tx *gorm.DB) error {
	if c.Balance == 0 {
		c.Balance = 30 // Начальный баланс
	}
	return nil
}

// IsWaitingAnswer проверяет, ждет ли чат ответа на вопрос
func (c *Chat) IsWaitingAnswer() bool {
	if c.LastQuestionID == nil || c.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().Before(*c.ExpiresAt)
}

// Question представляет вопрос в квизе
type Question struct {
	ID                uint       `gorm:"primaryKey;column:id;default:nextval('questions_id_seq')" json:"id"`
	Text              string     `gorm:"column:text;type:text;not null" json:"text"`
	Answer            string     `gorm:"column:answer;type:text;not null" json:"answer"`
	Comment           *string    `gorm:"column:comment;type:text" json:"comment"`
	AuthorID          *uint      `gorm:"column:author_id" json:"author_id"`
	Author            *Chat      `gorm:"foreignKey:AuthorID;constraint:OnDelete:SET NULL" json:"author"`
	IsPublished       bool       `gorm:"column:is_published;default:false;not null" json:"is_published"`
	QuestionPicture   *Picture   `gorm:"foreignKey:QuestionPictureID;constraint:OnDelete:SET NULL" json:"question_picture"`
	QuestionPictureID *uint      `gorm:"column:question_picture_id" json:"question_picture_id"`
	AnswerPicture     *Picture   `gorm:"foreignKey:AnswerPictureID;constraint:OnDelete:SET NULL" json:"answer_picture"`
	AnswerPictureID   *uint      `gorm:"column:answer_picture_id" json:"answer_picture_id"`
	ApprovedAt        *time.Time `gorm:"column:approved_at" json:"approved_at"`
	Rating            *int       `gorm:"column:rating;default:0" json:"rating"`
}

// TableName возвращает имя таблицы для Question
func (Question) TableName() string {
	return "questions"
}

// BeforeCreate хук перед созданием записи
func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.Rating == nil {
		zero := 0
		q.Rating = &zero
	}
	return nil
}

// Picture представляет изображение
type Picture struct {
	ID        uint      `gorm:"primaryKey;column:id;default:nextval('pictures_id_seq')" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Path      *string   `gorm:"column:path" json:"path"`
}

// TableName возвращает имя таблицы для Picture
func (Picture) TableName() string {
	return "pictures"
}

// Feedback представляет обратную связь
type Feedback struct {
	ID        uint      `gorm:"primaryKey;column:id;default:nextval('feedbacks_id_seq')" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Text      string    `gorm:"column:text;type:text;not null" json:"text"`
	Response  *string   `gorm:"column:response;type:text" json:"response"`
	ChatID    uint      `gorm:"column:chat_id;not null" json:"chat_id"`
	Chat      Chat      `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE" json:"chat"`
}

// TableName возвращает имя таблицы для Feedback
func (Feedback) TableName() string {
	return "feedbacks"
}

// Reaction представляет реакцию на вопрос
type Reaction struct {
	ID          uint       `gorm:"primaryKey;column:id;default:nextval('reactions_id_seq')" json:"id"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ResponsedAt *time.Time `gorm:"column:responsed_at" json:"responsed_at"`
	SkippedAt   *time.Time `gorm:"column:skipped_at" json:"skipped_at"`
	FailedAt    *time.Time `gorm:"column:failed_at" json:"failed_at"`
	ChatID      uint       `gorm:"column:chat_id;not null" json:"chat_id"`
	Chat        Chat       `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE" json:"chat"`
	QuestionID  uint       `gorm:"column:question_id;not null" json:"question_id"`
	Question    Question   `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"question"`
}

// TableName возвращает имя таблицы для Reaction
func (Reaction) TableName() string {
	return "reactions"
}

// BeforeCreate хук перед созданием записи
func (r *Reaction) BeforeCreate(tx *gorm.DB) error {
	// Уникальный индекс для комбинации chat_id и question_id
	tx.Statement.AddClause(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chat_id"}, {Name: "question_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"responsed_at", "skipped_at", "failed_at"}),
	})
	return nil
}
