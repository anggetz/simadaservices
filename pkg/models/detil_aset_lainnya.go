package models

type DetilAsetLainnya struct {
	ID            int    `json:"id"`
	Pidinventaris int    `json:"pidinventaris" `
	SeniPencipta  string `json:"seni_penciptas"`
	BukuJudul     string `json:"buku_judul"`
	TernakJenis   string `json:"ternak_jenis"`
}

func (i *DetilAsetLainnya) TableName() string {
	return "detil_aset_lainnya"
}
