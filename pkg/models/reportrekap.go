package models

import "time"

type Reportrekap struct {
	ID                            uint      `gorm:"column:id" json:"id"`
	NamaBarang                    string    `gorm:"column:nama_barang" json:"nama_barang"`
	TglDibukukan                  time.Time `gorm:"column:tgl_dibukukan" json:"tgl_dibukukan"`
	PidOPD                        uint      `gorm:"column:pidopd" json:"pidopd"`
	PidOPDCabang                  uint      `gorm:"column:pidopd_cabang" json:"pidopd_cabang"`
	PidUPT                        uint      `gorm:"column:pidupt" json:"pidupt"`
	KodeJenis                     string    `gorm:"column:kode_jenis" json:"kode_jenis"`
	KodeObjek                     string    `gorm:"column:kode_objek" json:"kode_objek"`
	KodeRincianObjek              string    `gorm:"column:kode_rincian_objek" json:"kode_rincian_objek"`
	KodeSubRincianObjek           string    `gorm:"column:kode_sub_rincian_objek" json:"kode_sub_rincian_objek"`
	KodeSubSubRincianObjek        string    `gorm:"column:kode_sub_sub_rincian_objek" json:"kode_sub_sub_rincian_objek"`
	Jumlah                        int       `gorm:"column:jumlah" json:"jumlah"`
	HargaSatuan                   float64   `gorm:"column:harga_satuan" json:"harga_satuan"`
	NilaiPerolehan                float64   `gorm:"column:nilai_perolehan" json:"nilai_perolehan"`
	Jenis                         uint      `gorm:"column:jenis" json:"jenis"`
	NilaiAtribusi                 float64   `gorm:"column:nilai_atribusi" json:"nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64   `gorm:"column:nilai_perolehan_setelah_atribusi" json:"nilai_perolehan_setelah_atribusi"`
	PidBarang                     uint      `gorm:"column:pidbarang" json:"pidbarang"`
	Draft                         string    `gorm:"column:draft" json:"draft"`
}

func (i *Reportrekap) TableName() string {
	return "public.reportrekap"
}
