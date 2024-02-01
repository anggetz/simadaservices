package models

type DetilKonstruksi struct {
	ID            int    `json:"id"`
	Pidinventaris int    `json:"pidinventaris"`
	Alamat        string `json:"alamat"`
}

func (i *DetilKonstruksi) TableName() string {
	return "detil_konstruksi"
}
