package models

import "time"

type MSatuanBarang struct {
	ID         uint       `gorm:"primaryKey"`
	Nama       string     `gorm:"column:nama"`
	Aktif      int        `gorm:"column:aktif"`
	BisaDibagi int        `gorm:"column:bisadibagi"`
	CreatedAt  *time.Time `gorm:"column:created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at"`
}

func (m *MSatuanBarang) TableName() string {
	return "m_satuan_barang"
}
