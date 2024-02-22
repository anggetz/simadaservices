package models

import "time"

type TaskQueue struct {
	ID           uint      `gorm:"primaryKey"`
	TaskName     string    `gorm:"type:varchar(255);not null"`
	TaskType     string    `gorm:"type:varchar(255);not null"`
	Status       string    `gorm:"type:varchar(255);not null"`
	CallbackLink string    `gorm:"type:varchar(255)"`
	CreatedAt    time.Time `gorm:"index"`
	UpdatedAt    time.Time `gorm:"index"`
	CreatedBy    int       // Assuming the ID of the user who created the task
	TaskUUID     string    `gorm:"type:varchar(255)"`
	TaskError    string    `gorm:"type:varchar(255)"`
}

func (t *TaskQueue) TableName() string {
	return "task_queue"
}
