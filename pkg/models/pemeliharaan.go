package models

import "time"

type Pemeliharaan struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	PIDInventaris      int        `json:"pidinventaris" gorm:"column:pidinventaris"`
	Tgl                *time.Time `json:"tgl"`
	Uraian             string     `json:"uraian"`
	Persh              string     `json:"persh"`
	Alamat             string     `json:"alamat"`
	NoKontrak          string     `json:"nokontrak"`
	TglKontrak         *time.Time `json:"tglkontrak"`
	Biaya              float64    `json:"biaya"`
	Menambah           int        `json:"menambah"`
	Keterangan         string     `json:"keterangan"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CreatedAt          time.Time  `json:"created_at"`
	Draft              string     `json:"draft"`
	CreatedBy          int        `json:"created_by"`
	IsExecByPenyusutan bool       `json:"is_exec_by_penyusutan"`
}

// Define the table name for the Pemeliharaan model
func (p *Pemeliharaan) TableName() string {
	return "pemeliharaan"
}
