package models

type Alamat struct {
	ID   int    `json:"id"`
	Nama string `json:"nama"`
}

func (i *Alamat) TableName() string {
	return "m_alamat"
}
