package models

type MerkBarang struct {
	ID   int    `json:"id"`
	Nama string `json:"nam"`
}

func (i *MerkBarang) TableName() string {
	return "m_merk_barang"
}
