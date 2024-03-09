package models

type Elastic struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore interface{} `json:"max_score"`
		Hits     []struct {
			Index  string                 `json:"_index"`
			Type   string                 `json:"_type"`
			ID     string                 `json:"_id"`
			Score  interface{}            `json:"_score"`
			Source map[string]interface{} `json:"_source"`
			// struct {
			// 	ID                  int         `json:"id"`
			// 	Noreg               string      `json:"noreg"`
			// 	Pidbarang           int         `json:"pidbarang"`
			// 	Pidopd              int         `json:"pidopd"`
			// 	Pidlokasi           int         `json:"pidlokasi"`
			// 	TglSensus           time.Time   `json:"tgl_sensus"`
			// 	Volume              int         `json:"volume"`
			// 	Pembagi             int         `json:"pembagi"`
			// 	HargaSatuan         int         `json:"harga_satuan"`
			// 	Perolehan           string      `json:"perolehan"`
			// 	Kondisi             string      `json:"kondisi"`
			// 	LokasiDetil         string      `json:"lokasi_detil"`
			// 	Keterangan          string      `json:"keterangan"`
			// 	UpdatedAt           time.Time   `json:"updated_at"`
			// 	CreatedAt           interface{} `json:"created_at"`
			// 	TahunPerolehan      string      `json:"tahun_perolehan"`
			// 	Jumlah              int         `json:"jumlah"`
			// 	TglDibukukan        time.Time   `json:"tgl_dibukukan"`
			// 	Satuan              int         `json:"satuan"`
			// 	DeletedAt           interface{} `json:"deleted_at"`
			// 	PidopdCabang        int         `json:"pidopd_cabang"`
			// 	PidUpt              int         `json:"pid_upt"`
			// 	KodeLokasi          string      `json:"kode_lokasi"`
			// 	AlamatPropinsi      int         `json:"alamat_propinsi"`
			// 	AlamatKota          int         `json:"alamat_kota"`
			// 	AlamatKecamatan     int         `json:"alamat_kecamatan"`
			// 	AlamatKelurahan     int64       `json:"alamat_kelurahan"`
			// 	Idpegawai           int         `json:"idpegawai"`
			// 	PidOrganisasi       int         `json:"pid_organisasi"`
			// 	KodeGedung          string      `json:"kode_gedung"`
			// 	KodeRuang           string      `json:"kode_ruang"`
			// 	PenanggungJawab     string      `json:"penanggung_jawab"`
			// 	UmurEkonomis        int         `json:"umur_ekonomis"`
			// 	Draft               string      `json:"draft"`
			// 	KodeBarang          string      `json:"kode_barang"`
			// 	ImportFlag          string      `json:"import_flag"`
			// 	NamaPopuler         string      `json:"nama_populer"`
			// 	IDSensus            int         `json:"id_sensus"`
			// 	TglPerolehan        time.Time   `json:"tgl_perolehan"`
			// 	IDPublish           int         `json:"id_publish"`
			// 	KodeNibar           string      `json:"kode_nibar"`
			// 	Jalan               string      `json:"jalan"`
			// 	Rt                  string      `json:"rt"`
			// 	Rw                  string      `json:"rw"`
			// 	VerifikatorFlag     bool        `json:"verifikator_flag"`
			// 	PostingFlag         bool        `json:"posting_flag"`
			// 	Noref               string      `json:"noref"`
			// 	VerifikatorStatus   int         `json:"verifikator_status"`
			// 	VerifikatorIsRevise bool        `json:"verifikator_is_revise"`
			// 	VerifikatorReviseBy int         `json:"verifikator_revise_by"`
			// }

			Sort []int64 `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}
