package usecase

import (
	"fmt"
	"libcore/models"
	"log"
	"time"

	"github.com/adjust/rmq/v5"
	"gorm.io/gorm"
)

type workerUseCase struct {
	db         *gorm.DB
	connection rmq.Connection
}

func WorkerUseCase(db *gorm.DB, connection rmq.Connection) *workerUseCase {
	return &workerUseCase{
		db:         db,
		connection: connection,
	}
}

func (w *workerUseCase) ReminderPenggunaanSementara() {
	// 3 bulan, 1 bulan, 1 minggu, dan berakhir
	penggunaan := []models.PenggunaanSementaraKurangDari{}
	if err := w.db.Where("state = ?", 3).Find(&penggunaan).Error; err != nil {
		log.Println("no data penggunaan found ")
		return
	}

	dateNow := time.Now().Local()
	var tglakhir time.Time
	title := "Penggunaan Sementara"

	for _, peng := range penggunaan {
		if peng.DokumenPengakhiranPemberiTanggal != nil {
			tglakhir = *peng.DokumenPengakhiranPemberiTanggal
		} else if peng.DokumenPengakhiranPemohonTanggal != nil {
			tglakhir = *peng.DokumenPengakhiranPemohonTanggal
		} else {
			tglakhir = *peng.TanggalBerakhir
		}

		if tglakhir.After(dateNow) {
			diff := dateNow.Sub(tglakhir)
			month := int(diff.Hours() / (24 * 30))
			week := int(diff.Hours() / (24 * 7))
			users := []models.User{}
			link_to := "penggunaan_sementara_kurang_dari/" + peng.ID
			sender := peng.OpdPemohonId

			// opd
			w.db.Where("pid_organisasi in (?,?)", peng.OpdPemilikId, peng.OpdPemohonId).Find(&users)
			idUsers := []int{}
			var body string
			for _, v := range users {
				idUsers = append(idUsers, v.ID)
			}

			// pengelola
			pengelola := []models.User{}
			w.db.Joins("m_organisasi mo on mo.id = users.pid_organisasi").
				Where("mo.level = ?", 0).
				Find(&pengelola)
			for _, v := range pengelola {
				idUsers = append(idUsers, v.ID)
			}

			if month == 3 || month == 1 {
				body = fmt.Sprintf(`Notifikasi Penggunaan Sementara Akan berakhir dalam %v bulan`, month)
			}
			if week == 1 {
				body = fmt.Sprintf(`Notifikasi Penggunaan Sementara Akan berakhir dalam %v minggu`, week)
			}
			if dateNow == tglakhir {
				body = `Notifikasi Penggunaan Sementara Sudah Berakhir`
			}

			err := CreateNotif(w.db, sender, idUsers, title, body, link_to)
			if err != nil {
				log.Println("error create notif id ", peng.ID, err.Error())
			}
		} else {
			log.Println("penggunaan sekarang berakhir", peng.ID)
		}
	}

	return
}
