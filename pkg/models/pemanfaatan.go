package models

import "time"

type Pemanfaatan struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	PIDInventaris    int        `json:"pidinventaris" gorm:"column:pidinventaris"`
	Peruntukan       string     `json:"peruntukan"`
	Umur             int        `json:"umur"`
	UmurSatuan       string     `json:"umur_satuan"`
	NoPerjanjian     string     `json:"no_perjanjian"`
	TglMulai         *time.Time `json:"tgl_mulai"`
	TglAkhir         *time.Time `json:"tgl_akhir"`
	Mitra            int        `json:"mitra"`
	TipeKontribusi   string     `json:"tipe_kontribusi"`
	JumlahKontribusi string     `json:"jumlah_kontribusi"`
	Aktif            string     `json:"aktif"`
	Pegawai          int        `json:"pegawai"`
	BagiHasil        int        `json:"bagi_hasil"`
	Tetap            int        `json:"tetap"`
	Draft            string     `json:"draft"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CreatedAt        time.Time  `json:"created_at"`
	CreatedBy        int        `json:"created_by"`
}

// Define the table name for the Pemeliharaan model
func (p *Pemanfaatan) TableName() string {
	return "pemanfataan"
}
