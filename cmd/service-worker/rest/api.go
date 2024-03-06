package rest

import (
	"simadaservices/cmd/service-report/rest"
	"simadaservices/pkg/usecase"

	"github.com/adjust/rmq/v5"
	"gorm.io/gorm"
)

type Api struct{}

func NewApi() *Api {
	return &Api{}
}

func (a *Api) GetBmdAtl(dbConnection *gorm.DB, connection rmq.Connection) {

	params := make(map[string]interface{})
	params["f_periode"] = "5"
	params["f_penggunafilter"] = "1" // dinas pendidikan
	params["penggunafilter"] = "1"
	params["f_kuasapengguna_filter"] = ""
	params["kuasapengguna_filter"] = ""
	params["f_subkuasa_filter"] = ""
	params["subkuasa_filter"] = ""
	params["f_tahun"] = "2024"
	params["f_bulan"] = "12"
	params["f_jenis"] = "Intrakomptabel"
	params["action"] = "export"
	params["firstload"] = "true"
	params["draw"] = "1"

	rest.NewApi().CronExportBmdAtl(dbConnection, connection, params)
}

func (a *Api) GetReminderPenggunaanSementara(dbConnection *gorm.DB, connection rmq.Connection) {

	usecase.WorkerUseCase(dbConnection, connection).ReminderPenggunaanSementara()

}
