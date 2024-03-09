package models

type Organisasi struct {
	ID    int    `json:"id"`
	PID   int    `json:"pid"`
	Nama  string `json:"nama"`
	Level int    `json:"level"`
}

func (o *Organisasi) TableName() string {
	return "m_organisasi"
}
