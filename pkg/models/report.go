package models

type ResponseInventaris struct {
	*Inventaris
	NilaiPerolehan   float64 `json:"nilai_perolehan"`
	NamaRekAset      string  `json:"nama_barang"`
	KodeJenis        string  `json:"kode_jenis"`
	KodeObjek        string  `json:"kode_objek"`
	KodeRincianObjek string  `json:"kode_rincian_objek"`
	OrganisasiNama   string  `json:"pengguna_barang"`
	StatusBarang     string  `json:"status_barang"`
	StatusApproval   string  `json:"status_approval"`
	TglWaktuSensus   string  `json:"tglwaktusensus"`
	StatusSensus     string  `json:"status_sensus"`
}

type SummaryPerPage struct {
	NilaiPerolehan                float64 `json:"nilai_perolehan"`
	NilaiAtribusi                 float64 `json:"nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64 `json:"nilai_perolehan_setelah_atribusi"`
	NilaiAkumulasiPenyusutan      float64 `json:"nilai_akumulasi_penyusutan"`
	NilaiBuku                     float64 `json:"nilai_buku"`
	NilaiHargaSatuan              float64 `json:"nilai_harga_satuan"`
	NilaiPenyusutanTahun          float64 `json:"nilai_penyusutan_sd_tahun"`
	NilaiPenyusutanPeriode        float64 `json:"nilai_penyusutan_sd_periode"`
	NilaiBebanPenyusutan          float64 `json:"nilai_beban_penyusutan"`
}

type SummaryPage struct {
	Jumlah                        int     `json:"jumlah"`
	NilaiHargaSatuan              float64 `json:"total_nilai_harga_satuan"`
	NilaiPerolehan                float64 `json:"total_nilai_perolehan"`
	NilaiAtribusi                 float64 `json:"total_nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64 `json:"total_nilai_perolehan_setelah_atribusi"`
	NilaiPenyusutanTahun          float64 `json:"total_nilai_penyusutan_sd_tahun"`
	NilaiPenyusutanPeriode        float64 `json:"total_nilai_penyusutan_sd_periode"`
	NilaiBebanPenyusutan          float64 `json:"total_nilai_beban_penyusutan"`
	NilaiBuku                     float64 `json:"total_nilai_buku"`
	NilaiAkumulasiPenyusutan      float64 `json:"total_nilai_akumulasi_penyusutan"`
}

type ResponseRekapitulasi struct {
	NamaBarang                    string  `json:"nama_barang"`
	KodeBarang                    string  `json:"kode_barang"`
	Jumlah                        int64   `json:"jumlah"`
	NilaiPerolehan                float64 `json:"nilai_perolehan"`
	NilaiAtribusi                 float64 `json:"nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64 `json:"nilai_perolehan_setelah_atribusi"`
	AkumulasiPenyusutan           float64 `json:"akumulasi_penyusutan"`
	NilaiBuku                     float64 `json:"nilai_buku"`
}

type ReportMDBATL struct {
	KodeAkun                      string  `json:"kode_akun"`
	KodeKelompok                  string  `json:"kode_kelompok"`
	KodeJenis                     string  `json:"kode_jenis"`
	KodeObjek                     string  `json:"kode_objek"`
	KodeRincianObjek              string  `json:"kode_rincian_objek"`
	KodeSubRincianObjek           string  `json:"kode_sub_rincian_objek"`
	KodeSubSubRincianObjek        string  `json:"kode_sub_sub_rincian_objek"`
	PIDOpd                        int     `json:"pidopd"`
	PIDOPDCabang                  int     `json:"pidopd_cabang"`
	PIDUpt                        int     `json:"pid_upt"`
	Nama                          string  `json:"nama"`
	Level                         int     `json:"level"`
	NamaBarang                    string  `json:"nama_barang"`
	Nibar                         string  `json:"nibar"`
	NomorRegister                 string  `json:"nomor_register"`
	SpesifikasiNamaBarang         string  `json:"spesifikasi_nama_barang"`
	SpesifikasiLainnya            string  `json:"spesifikasi_lainnya"`
	Lokasi                        string  `json:"lokasi"`
	Jumlah                        int     `json:"jumlah"`
	Satuan                        string  `json:"satuan"`
	HargaSatuanPerolehan          float64 `json:"harga_satuan_perolehan"`
	NilaiPerolehan                float64 `json:"nilai_perolehan"`
	NilaiAtribusi                 float64 `json:"nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64 `json:"nilai_perolehan_setelah_atribusi"`
	PenyusutanTahunSebelumnya     float64 `json:"penyusutan_sd_tahun_sebelumnya"`
	BebanPenyusutan               float64 `json:"beban_penyusutan"`
	PenyusutanTahunSekarang       float64 `json:"penyusutan_sd_tahun_sekarang"`
	NilaiBuku                     float64 `json:"nilai_buku"`
	CaraPerolehan                 string  `json:"cara_perolehan"`
	TglPerolehan                  string  `json:"tgl_perolehan"`
	StatusPenggunaan              string  `json:"status_penggunaan"`
	Keterangan                    string  `json:"keterangan"`
	TglDibukukan                  string  `json:"tgl_dibukukan"`
	Tahun                         int     `json:"tahun"`
	Bulan                         string  `json:"bulan"`
}

type FileStruct struct {
	FilePath  string
	FileName  string
	FileSize  float64
	CreatedAt string
}
