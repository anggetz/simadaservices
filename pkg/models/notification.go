package models

import "time"

type Notification struct {
	ID        int       `gorm:"primary_key"`
	CreatedBy int       // Sesuaikan dengan tipe data yang sesuai
	Body      string    `gorm:"type:varchar(500);not null"`
	Title     string    `gorm:"type:varchar(100);not null"`
	LinkTo    string    `gorm:"type:varchar(250);not null"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time `gorm:"index"`
}

func (t *Notification) TableName() string {
	return "notification"
}
