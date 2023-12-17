package models

type Organisasi struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Level int    `json:"id"`
}

func (o *Organisasi) TableName() string {
	return "m_organisasi"
}
