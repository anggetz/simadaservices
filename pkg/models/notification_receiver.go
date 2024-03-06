package models

import "time"

type NotificationReceiver struct {
	ID             int        `gorm:"primary_key"`
	NotificationID int        // Sesuaikan dengan tipe data yang sesuai
	Status         bool       // Sesuaikan dengan tipe data yang sesuai
	ViewAt         *time.Time `gorm:"type:timestamp(0)"`
	ReceiverID     int        // Sesuaikan dengan tipe data yang sesuai
	CreatedAt      time.Time  `gorm:"index"`
	UpdatedAt      time.Time  `gorm:"index"`
}

func (t *NotificationReceiver) TableName() string {
	return "notification_receiver"
}
