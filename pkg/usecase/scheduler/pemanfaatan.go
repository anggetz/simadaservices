package scheduler

import (
	"fmt"
	"simadaservices/pkg/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Pemanfaatan interface {
	Execute() error
}

type pemanfaatanImpl struct {
	db *gorm.DB
}

func NewPemanfaatan(db *gorm.DB) Pemanfaatan {
	return &pemanfaatanImpl{
		db: db,
	}
}

func (p *pemanfaatanImpl) Execute() error {
	pemanfaatans := []models.Pemanfaatan{}

	txDbPemanfaatan := p.db.Raw(`
		select * from pemanfaatan where TO_DATE(date_part('year', NOW()) || '-' || date_part('month', tgl_mulai) || '-' || date_part('day', tgl_mulai), 'YYYY-MM-DD') - INTERVAL '40 DAY' = DATE(NOW())
	`).Scan(&pemanfaatans)

	if txDbPemanfaatan.Error != nil {
		return txDbPemanfaatan.Error
	}

	orgs := []models.Organisasi{}
	txFetchOrg := p.db.Model(new(models.Organisasi)).Where("level = ?", -1).Find(&orgs)
	if txFetchOrg.Error != nil {
		return txFetchOrg.Error
	}

	users := []models.User{}
	// get all pengelola users
	for _, org := range orgs {
		userOrgs := []models.User{}
		p.db.Model(new(models.User)).Where("pid_organisasi = ?", org.ID).Find(&userOrgs)

		users = append(users, userOrgs...)
	}

	for _, pemanfaatan := range pemanfaatans {
		//  create notif
		tglMulaiWithYearNow := time.Date(pemanfaatan.TglMulai.Year(), pemanfaatan.TglMulai.Month(), pemanfaatan.TglMulai.Day(), pemanfaatan.TglMulai.Hour(), pemanfaatan.TglMulai.Minute(), pemanfaatan.TglMulai.Second(), 0, pemanfaatan.TglMulai.Location())

		notif := &models.Notification{
			Body: "Jatuh Tempo Pembayaran Kontribusi Tetap " + pemanfaatan.NoPerjanjian + " mitra " + strconv.Itoa(pemanfaatan.Mitra) + " barang paling lambat pada " + tglMulaiWithYearNow.Format("2006-01-02 15:04:05") +
				" sebesar Rp. " + pemanfaatan.JumlahKontribusi,
			CreatedBy: users[0].ID,
			Title:     "Info jatuh tempo",
			LinkTo:    "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		txCreateNotif := p.db.Save(&notif)
		if txCreateNotif.Error != nil {
			fmt.Println("here!", len(users))
			return txCreateNotif.Error
		}

		fmt.Println("here!", len(users))

		for _, usr := range users {

			// receiver
			notifReceiver := &models.NotificationReceiver{
				NotificationID: notif.ID,
				Status:         false,
				ViewAt:         nil,
				ReceiverID:     usr.ID,
				CreatedAt:      time.Now(),
			}

			txCreateReceiver := p.db.Create(&notifReceiver)
			if txCreateReceiver.Error != nil {
				return txCreateReceiver.Error
			}
		}

	}
	return nil
}
