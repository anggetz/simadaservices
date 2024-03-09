package models

type DetilJalan struct {
	ID            int    `json:"id"`
	Pidinventaris int    `json:"pidinventaris"`
	Alamat        string `json:"alamat"`
	KodeJalan     string `json:"kode_jalan"`
}

func (i *DetilJalan) TableName() string {
	return "detil_jalan"
}
