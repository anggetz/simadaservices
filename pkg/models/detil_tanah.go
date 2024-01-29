package models

type DetilTanah struct {
	ID            int     `json:"id" gorm:"column:id;primaryKey"`
	Pidinventaris int     `json:"pidinventaris" gorm:"column:pidinventaris"`
	Luas          float64 `json:"luas"`
	Alamat        string  `json:"alamat"`
	// Idkota           int         `json:"idkota"`
	// Idkecamatan      int         `json:"idkecamatan"`
	// Idkelurahan      int         `json:"idkelurahan"`
	// Koordinatlokasi  string      `json:"koordinat_lokasi"`
	// Koordinattanah   string      `json:"koordinat_tanah"`
	// Hak              string      `json:"hak"`
	StatusSertifikat string `json:"status_sertifikat"`
	// TglSertifikat    *time.Time  `json:"tgl_sertifikat"`
	NomorSertifikat string `json:"nomor_sertifikat"`
	Penggunaan      string `json:"penggunaan"`
	// Keterangan       string      `json:"keterangan"`
	// Dokumen          string      `json:"dokumen"`
	// Foto             string      `json:"foto"`
	// CreatedAt        *time.Time  `json:"created_at"`
	// UpdatedAt        *time.Time  `json:"updated_at"`
	// NilaiHub         int         `json:"nilai_hub"`
	// Tipe             string      `json:"tipe"`
	// Geom             interface{} `json:"geom"`
	// BatasUtara       string      `json:"batas_utara"`
	// BatasTimur       string      `json:"batas_timur"`
	// Batasbarat       string      `json:"batas_barat"`
	// BatasSelatan     string      `json:"batas_selatan"`
}

func (i *DetilTanah) TableName() string {
	return "detil_tanah"
}
