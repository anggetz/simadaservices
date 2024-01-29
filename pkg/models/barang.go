package models

import (
	"time"
)

type Barang struct {
	// gorm.Model
	ID                  int        `json:"id"`
	NamaRekAset         string     `json:"nama_rek_aset"`
	UpdatedAt           *time.Time `json:"updated_at"`
	CreatedAt           *time.Time `json:"created_at"`
	KodeAkun            string     `json:"kode_akun"`
	KodeKelompok        string     `json:"kode_kelompok"`
	KodeJenis           string     `json:"kode_jenis"`
	KodeObjek           string     `json:"kode_objek"`
	KodeRincianObjek    string     `json:"kode_rincian_objek"`
	KodeSubRincianObjek string     `json:"kode_sub_rincian_objek"`
	UmurEkonomis        int        `json:"umur_ekonomis"`
	IsUseDefaultForm    bool       `json:"is_use_default_form"`
	FormUse             string     `json:"form_use"`
	Aktif               bool       `json:"aktif"`
}

func (i *Barang) TableName() string {
	return "m_barang"
}
