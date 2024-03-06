package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    int    `gorm:"primaryKey"`
	Name                  string `gorm:"column:name"`
	Email                 string
	EmailVerifiedAt       *time.Time
	Password              string
	RememberToken         string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	NIP                   string
	NoHP                  string     `gorm:"column:no_hp"`
	TanggalLahir          *time.Time `gorm:"column:tgl_lahir"`
	JenisKelamin          string     `gorm:"column:jenis_kelamin"`
	PIDOrganisasi         int        `gorm:"column:pid_organisasi"`
	Role                  int
	Username              string
	Aktif                 string
	EmailVerificationCode string `gorm:"column:email_verification_code"`
	Jabatan               int
	APIToken              string     `gorm:"column:api_token"`
	EmailForgotPassword   string     `gorm:"column:email_forgot_password"`
	APIExpired            *time.Time `gorm:"column:api_expired"`
	APIIntegration        bool       `gorm:"column:api_integration"`
	APISimada             bool       `gorm:"column:api_simada"`
	UserGroupRoleID       uuid.UUID  `gorm:"column:user_group_role_id"`
}

func (i *User) TableName() string {
	return "users"
}
