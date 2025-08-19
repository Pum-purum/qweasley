package models

import (
	"time"

	"gorm.io/gorm"
)

// Chat представляет чат пользователя
type Chat struct {
	ID         uint           `gorm:"primaryKey;column:id;default:nextval('chats_id_seq')" json:"id"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	Balance    int            `gorm:"column:balance;default:0;not null" json:"balance"`
	TelegramID int64          `gorm:"column:telegram_id;uniqueIndex;not null" json:"telegram_id"`
	Title      *string        `gorm:"column:title" json:"title"`
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

// Question представляет вопрос в квизе
type Question struct {
	ID          uint           `gorm:"primaryKey;column:id;default:nextval('questions_id_seq')" json:"id"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	Text        string         `gorm:"column:text;type:text;not null" json:"text"`
	Answer      string         `gorm:"column:answer;type:text;not null" json:"answer"`
	Comment     *string        `gorm:"column:comment;type:text" json:"comment"`
	AuthorID    *uint          `gorm:"column:author_id" json:"author_id"`
	Author      *Chat          `gorm:"foreignKey:AuthorID;constraint:OnDelete:SET NULL" json:"author"`
	IsPublished bool           `gorm:"column:is_published;default:false;not null" json:"is_published"`
	ApprovedAt  *time.Time     `gorm:"column:approved_at" json:"approved_at"`
	Rating      *int           `gorm:"column:rating;default:0" json:"rating"`
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

// Feedback представляет обратную связь
type Feedback struct {
	ID        uint           `gorm:"primaryKey;column:id;default:nextval('feedbacks_id_seq')" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	Text      string         `gorm:"column:text;type:text;not null" json:"text"`
	Response  *string        `gorm:"column:response;type:text" json:"response"`
	ChatID    uint           `gorm:"column:chat_id;not null" json:"chat_id"`
	Chat      Chat           `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE" json:"chat"`
}

// TableName возвращает имя таблицы для Feedback
func (Feedback) TableName() string {
	return "feedbacks"
}
