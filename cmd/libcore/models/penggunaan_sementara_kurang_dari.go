package models

import "time"

type PenggunaanSementaraKurangDari struct {
	ID                                  string     `json:"id"`
	OpdPemilikId                        int        `json:"opd_pemilik_id"`
	OpdPemohonId                        int        `json:"opd_pemohon_id"`
	DokumenPermohonanNomor              string     `json:"dokumen_permohonan_nomor"`
	DokumenPerjanjianPemberiJangkawaktu int32      `json:"dokumen_perjanjian_pemberi_jangkawaktu"`
	State                               int        `json:"state"`
	Status                              string     `json:"status"`
	TanggalPersetujuan                  *time.Time `json:"tanggal_persetujuan"`
	TanggalBerakhir                     *time.Time `json:"tanggal_berakhir"`
	DokumenPengakhiranPemberiTanggal    *time.Time `json:"dokumen_pengakhiran_pemberi_tanggal"`
	DokumenPengakhiranPemohonTanggal    *time.Time `json:"dokumen_pengakhiran_pemohon_tanggal"`
}

func (i *PenggunaanSementaraKurangDari) TableName() string {
	return "penggunaan_sementara_kurang_dari"
}
