package models

type InventarisSaldoAwal struct {
	ID     int     `json:"id"`
	Pidopd int     `json:"pidopd"`
	Nilai  float64 `json:"nilai"`
	Tipe   string  `json:"tipe"`
	Year   int     `json:"aktif"`
}

func (m *InventarisSaldoAwal) TableName() string {
	return "inventaris_saldoawal"
}
