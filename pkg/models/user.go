package models

type User struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	ApiToken      string `json:"api_token"`
	PidOrganisasi int    `json:"pid_organisasi" gorm:"column:pid_organisasi"`
	Level         int    `json:"level"`
}
