package models

type MMitra struct {
	ID      int    `json:"id"`
	NPWP    string `json:"npwp"`
	SiupTdp string `json:"siup_tdp" gorm:"column:siup_tdp"`
	Nama    string `json:"nama"`
	Alamat  string `json:"alamat"`
}

func (m *MMitra) TableName() string {
	return "m_mitra"
}
