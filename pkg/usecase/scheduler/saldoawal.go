package scheduler

import (
	"simadaservices/pkg/models"

	"gorm.io/gorm"
)

type saldoAwal struct {
	db *gorm.DB
}

type SaldoAwal interface {
	Execute(pidopd, year int) (bool, error)
}

func NewSaldoAwal(db *gorm.DB) SaldoAwal {
	return &saldoAwal{
		db: db,
	}
}

type saldoAwalExe struct {
	SaldoAwal float64
	PidOpd    int
}

func (s *saldoAwal) Execute(pidopd, year int) (bool, error) {

	data := saldoAwalExe{}
	dataPrev := saldoAwalExe{}

	// check if previous year is exists or not
	txDbPrev := s.db.Raw(`
		select sum(harga_satuan) as saldo_awal, pidopd from inventaris where date_part('year', tgl_dibukukan) = ? and pidopd = ? group by date_part('year', tgl_dibukukan), pidopd
	`, year-2, pidopd).Scan(&dataPrev)

	if txDbPrev.Error != nil {
		return false, txDbPrev.Error
	}

	operatorYear := "="

	if txDbPrev.Error != gorm.ErrRecordNotFound {
		operatorYear = "<="
	}

	txDb := s.db.Raw(`
		select sum(harga_satuan) as saldo_awal, pidopd from inventaris where date_part('year', tgl_dibukukan) `+operatorYear+` ? and pidopd = ? group by date_part('year', tgl_dibukukan), pidopd
	`, year-1, pidopd).Scan(&data)

	if txDb.Error != nil {
		return false, txDb.Error
	}

	saldoAwal := models.InventarisSaldoAwal{
		Pidopd: pidopd,
		Nilai:  float64(data.SaldoAwal),
		Tipe:   "Saldo Awal",
		Year:   year,
	}

	// save to saldo awal report
	txCreate := s.db.Create(&saldoAwal)

	if txCreate.Error != nil {
		return false, txCreate.Error
	}

	return true, nil
}
