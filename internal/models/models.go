package models

import (
	"time"

	"gorm.io/gorm"
)

// Chat представляет чат пользователя
type Chat struct {
	ID         uint           `gorm:"primaryKey;column:id;default:nextval('chats_id_seq')" json:"id"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Balance    int            `gorm:"column:balance;default:0" json:"balance"`
	TelegramID int64          `gorm:"column:telegram_id;uniqueIndex;not null" json:"telegram_id"`
	Title      *string        `gorm:"column:title" json:"title"`
}

// TableName возвращает имя таблицы для Chat
func (Chat) TableName() string {
	return "chats"
}

// Question представляет вопрос в квизе
type Question struct {
	ID                uint           `gorm:"primaryKey;column:id;default:nextval('questions_id_seq')" json:"id"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	Text              string         `gorm:"column:text;type:text" json:"text"`
	Answer            string         `gorm:"column:answer;type:text" json:"answer"`
	Comment           *string        `gorm:"column:comment;type:text" json:"comment"`
	AuthorID          *uint          `gorm:"column:author_id" json:"author_id"`
	Author            *Chat          `gorm:"foreignKey:AuthorID" json:"author"`
	IsPublished       bool           `gorm:"column:is_published;default:false" json:"is_published"`
	QuestionPicture   *Picture       `gorm:"foreignKey:QuestionPictureID" json:"question_picture"`
	QuestionPictureID *uint          `gorm:"column:question_picture_id" json:"question_picture_id"`
	AnswerPicture     *Picture       `gorm:"foreignKey:AnswerPictureID" json:"answer_picture"`
	AnswerPictureID   *uint          `gorm:"column:answer_picture_id" json:"answer_picture_id"`
	ApprovedAt        *time.Time     `gorm:"column:approved_at" json:"approved_at"`
	Rating            *int           `gorm:"column:rating;default:0" json:"rating"`
}

// TableName возвращает имя таблицы для Question
func (Question) TableName() string {
	return "questions"
}

// Picture представляет изображение
type Picture struct {
	ID        uint           `gorm:"primaryKey;column:id;default:nextval('pictures_id_seq')" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Path      *string        `gorm:"column:path" json:"path"`
}

// TableName возвращает имя таблицы для Picture
func (Picture) TableName() string {
	return "pictures"
}

// Feedback представляет обратную связь
type Feedback struct {
	ID        uint           `gorm:"primaryKey;column:id;default:nextval('feedbacks_id_seq')" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Text      string         `gorm:"column:text;type:text" json:"text"`
	Response  *string        `gorm:"column:response;type:text" json:"response"`
	ChatID    uint           `gorm:"column:chat_id;not null" json:"chat_id"`
	Chat      Chat           `gorm:"foreignKey:ChatID" json:"chat"`
}

// TableName возвращает имя таблицы для Feedback
func (Feedback) TableName() string {
	return "feedbacks"
}

// Reaction представляет реакцию на вопрос
type Reaction struct {
	ID         uint           `gorm:"primaryKey;column:id;default:nextval('reactions_id_seq')" json:"id"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	ChatID     uint           `gorm:"column:chat_id;not null" json:"chat_id"`
	Chat       Chat           `gorm:"foreignKey:ChatID" json:"chat"`
	QuestionID uint           `gorm:"column:question_id;not null" json:"question_id"`
	Question   Question       `gorm:"foreignKey:QuestionID" json:"question"`
	IsCorrect  bool           `gorm:"column:is_correct" json:"is_correct"`
}

// TableName возвращает имя таблицы для Reaction
func (Reaction) TableName() string {
	return "reactions"
}
