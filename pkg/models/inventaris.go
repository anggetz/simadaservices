package models

import (
	"time"
)

type Inventaris struct {
	ID                  int               `json:"id"`
	Noreg               string            `json:"noreg"`
	Pidbarang           int               `json:"pidbarang"`
	PIDOpd              int               `json:"pidopd"`
	PIDLokasi           int               `json:"pidlokasi"`
	TglSensus           *time.Time        `json:"tgl_sensus"`
	Volume              int               `json:"volume"`
	Pembagi             int               `json:"pembagi"`
	HargaSatuan         float64           `json:"harga_satuan"`
	Perolehan           string            `json:"perolehan" gorm:"column:perolehan"`
	Kondisi             string            `json:"kondisi"`
	LokasiDetil         string            `json:"lokasi_detil"`
	Keterangan          string            `json:"keterangan"`
	UpdatedAt           *time.Time        `json:"updated_at"`
	CreatedAt           *time.Time        `json:"created_at"`
	TahunPerolehan      string            `json:"tahun_perolehan"`
	Jumlah              int               `json:"jumlah"`
	TglDibukukan        *time.Time        `json:"tgl_dibukukan"`
	Satuan              int               `json:"satuan"`
	DeletedAt           *time.Time        `json:"deleted_at"`
	PIDOPDCabang        int               `json:"pidopd_cabang"`
	PIDUpt              int               `json:"pid_upt"`
	KodeLokasi          string            `json:"kode_lokasi"`
	AlamatPropinsi      int               `json:"alamat_propinsi"`
	AlamatKota          int               `json:"alamat_kota"`
	AlamatKecamatan     int               `json:"alamat_kecamatan"`
	AlamatKelurahan     int               `json:"alamat_kelurahan"`
	Idpegawai           int               `json:"idpegawai"`
	PidOrganisasi       int               `json:"pid_organisasi"`
	KodeGedung          string            `json:"kode_gedung"`
	KodeRuang           string            `json:"kode_ruang"`
	PenanggungJawab     string            `json:"penanggung_jawab"`
	UmurEkonomis        int               `json:"umur_ekonomis"`
	Draft               string            `json:"draft"`
	KodeBarang          string            `json:"kode_barang"`
	ImportFlag          string            `json:"import_flag"`
	NamaPopuler         string            `json:"nama_populer"`
	IdSensus            int               `json:"id_sensus"`
	TglPerolehan        *time.Time        `json:"tgl_perolehan"`
	IdPublish           int               `json:"id_publish"`
	KodeNibar           string            `json:"kode_nibar"`
	Jalan               string            `json:"jalan"`
	RT                  string            `json:"rt"`
	RW                  string            `json:"rw"`
	VerifikatorFlag     bool              `json:"verifikator_flag"`
	PostingFlag         bool              `json:"posting_flag"`
	Noref               string            `json:"noref"`
	VerifikatorStatus   int               `json:"verifikator_status"`
	VerifikatorIsRevise bool              `json:"verifikator_is_revise"`
	VerifikatorReviseBy int64             `json:"verifikator_revise_by"`
	BarangMaster        *Barang           `json:"barang" gorm:"foreignKey:pidbarang"`
	DetilTanahRel       *DetilTanah       `json:"detil_tanah" gorm:"foreignKey:ID;references:pidinventaris"`
	DetilMesinRel       *DetilMesin       `json:"detil_mesin" gorm:"foreignKey:ID;references:pidinventaris"`
	DetilAsetLainnyaRel *DetilAsetLainnya `json:"detil_aset_lainnya" gorm:"foreignKey:ID;references:pidinventaris"`
	DetilBangunanRel    *DetilBangunan    `json:"detil_bangunan" gorm:"foreignKey:ID;references:pidinventaris"`
	DetilJalanRel       *DetilJalan       `json:"detil_jalan" gorm:"foreignKey:ID;references:pidinventaris"`
	DetilKonstruksiRel  *DetilKonstruksi  `json:"detil_konstruksi" gorm:"foreignKey:ID;references:pidinventaris"`
	AlamatKotaRel       *Alamat           `json:"alamat_kota_rel" gorm:"foreignKey:alamat_kota;references:id"`
	AlamatKecamatanRel  *Alamat           `json:"alamat_kecamatan_rel" gorm:"foreignKey:alamat_kecamatan;references:id"`
}

func (i *Inventaris) TableName() string {
	return "inventaris"
}
