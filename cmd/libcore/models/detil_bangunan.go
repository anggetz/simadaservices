package models

type DetilBangunan struct {
	ID            int     `json:"id"`
	Pidinventaris int     `json:"pidinventaris" `
	Alamat        string  `json:"alamat"`
	Luas          float64 `json:"luas"`
}

func (i *DetilBangunan) TableName() string {
	return "detil_bangunan"
}
