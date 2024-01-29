package models

import "time"

type DetilAsetLainnya struct {
	ID                                uint       `json:"id"`
	PidInventaris                     int        `json:"pidinventaris"`
	BukuJudul                         string     `json:"buku_judul"`
	BukuSpesifikasi                   string     `json:"buku_spesifikasi"`
	SeniAsal                          string     `json:"seni_asal"`
	SeniPencipta                      string     `json:"seni_pencipta"`
	SeniBahan                         string     `json:"seni_bahan"`
	TernakJenis                       string     `json:"ternak_jenis"`
	TernakUkuran                      string     `json:"ternak_ukuran"`
	Keterangan                        string     `json:"keterangan"`
	Dokumen                           string     `json:"dokumen"`
	Foto                              string     `json:"foto"`
	NamaDokumenKepemilikan            string     `json:"nama_dokumen_kepemilikan"`
	NomorDokumenKepemilikan           string     `json:"nomor_dokumen_kepemilikan"`
	TanggalDokumenKepemilikan         *time.Time `json:"tanggal_dokumen_kepemilikan"`
	NamaKepemilikanDokumenKepemilikan string     `json:"nama_kepemilikan_dokumen_kepemilikan"`
	SpesifikasiNamaBarang             string     `json:"spesifikasi_nama_barang"`
	SpesifikasiLainnya                string     `json:"spesifikasi_lainnya"`
}

func (d *DetilAsetLainnya) TableName() string {
	return "detil_aset_lainnya"
}
