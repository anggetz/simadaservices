package models

type DetilMesin struct {
	ID            int         `json:"id"`
	Pidinventaris int         `json:"pidinventaris"`
	Norangka      string      `json:"norangka"`
	Nomesin       string      `json:"nomesin"`
	Nopol         string      `json:"nopol"`
	Merk          string      `json:"merk"`
	MerkMaster    *MerkBarang `json:"merk_master" gorm:"foreignKey:Merk"`
}

func (i *DetilMesin) TableName() string {
	return "detil_mesin"
}
