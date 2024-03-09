package rest

import (
	"fmt"
	"libcore/tools"
	"libcore/usecase"
	"service-transaction/kernel"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type InvoiceApi interface {
	Get(gin *gin.Context)
	GetQueueExportInventaris(gin *gin.Context)
	GetInventarisPemeliharaan(gin *gin.Context)
	GetInventarisNeedVerification(gin *gin.Context)
}

type InvoiceImpl struct {
}

func NewInvoiceApi() InvoiceApi {
	return &InvoiceImpl{}
}

func (a *InvoiceImpl) GetQueueExportInventaris(g *gin.Context) {
	connectionRedis := *kernel.Kernel.Config.REDIS.Connection
	queues, err := connectionRedis.GetOpenQueues()
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	for _, q := range queues {
		// queue, err := connectionRedis.OpenQueue(q)
		// if err != nil {
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

		statQ, err := connectionRedis.CollectStats([]string{q})
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		fmt.Println("ready count worker:", statQ.QueueStats["excel-worker"].UnackedCount())
	}

	// return g.JSON(preQueueWorkerExcel.)

}

func (a *InvoiceImpl) Get(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))
	action := g.Query("action")

	t, _ := g.Get("token_info")

	// fmt.Println(limit, page)
	if action == "export-excel" {
		// connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		// preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_EXPORT_EXCEL_INVENTARIS)
		// if err != nil {
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

		// // insert record into task_queue table
		// tq, err := usecase.
		// 	NewInventarisUseCase().
		// 	SetDB(kernel.Kernel.Config.DB.Connection).
		// 	SetRedisCache(kernel.Kernel.Config.REDIS.Cache).
		// 	SetRegisterQueue(g)

		// if err != nil {
		// 	log.Println("Error insert task queue: ", err.Error())
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

		// payload := map[string]interface{}{
		// 	"published":                         g.Query("published"),
		// 	"except-id-inventaris":              g.Query("except-id-inventaris"),
		// 	"pencarian_khusus":                  g.Query("pencarian_khusus"),
		// 	"pencarian_khusus_nilai":            g.Query("pencarian_khusus_nilai"),
		// 	"pencarian_khusus_range":            g.Query("pencarian_khusus_range"),
		// 	"pencarian_khusus_range_nilai_from": g.Query("pencarian_khusus_range_nilai_from"),
		// 	"pencarian_khusus_range_nilai_to":   g.Query("pencarian_khusus_range_nilai_to"),
		// 	"jenisbarangs":                      g.Query("jenisbarangs"),
		// 	"kodeobjek":                         g.Query("kodeobjek"),
		// 	"koderincianobjek":                  g.Query("koderincianobjek"),
		// 	"penggunafilter":                    g.Query("penggunafilter"),
		// 	"kuasapengguna_filter":              g.Query("kuasapengguna_filter"),
		// 	"subkuasa_filter":                   g.Query("subkuasa_filter"),
		// 	"draft":                             g.Query("draft"),
		// 	"status_sensus":                     g.Query("status_sensus"),
		// 	"status_verifikasi":                 g.Query("status_verifikasi"),
		// 	"queue_id":                          tq.ID,
		// 	"token_username":                    t.(jwt.MapClaims)["username"],
		// 	"token_org":                         t.(jwt.MapClaims)["org_id"],
		// 	"token_id":                          t.(jwt.MapClaims)["id"],
		// }

		// params, _ := json.Marshal(payload)
		// err = preQueueWorkerExcel.PublishBytes(params)
		// if err != nil {
		// 	log.Println("Error publish task queue: ", err.Error())
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

		// g.JSON(200, tools.HttpResponse{
		// 	Data:    tq,
		// 	Message: "Please wait! file is exporting.",
		// })
		// return
	} else {
		inventaris, totalFiltered, total, err := usecase.
			NewInventarisUseCase().
			SetDB(kernel.Kernel.Config.DB.Connection).
			SetRedisCache(kernel.Kernel.Config.REDIS.Cache).
			Get(
				length,
				start,
				usecase.NewAuthUseCase(kernel.Kernel.Config.DB.Connection).
					IsUserHasAccess(t.(jwt.MapClaims)["id"].(float64),
						[]string{"transaksi.inventaris.delete"}),
				g,
			)

		// inventaris, totalFiltered, total, err := usecase.
		// 	NewInventarisUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.ELASTIC.Client).
		// 	GetFromElastic(
		// 		length,
		// 		start,
		// 		g,
		// 	)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}
		g.JSON(200, tools.HttpResponse{
			Message:         "success get data",
			Data:            inventaris,
			RecordsFiltered: totalFiltered,
			RecordsTotal:    total,
		})
	}

	return
}

func (a *InvoiceImpl) GetInventarisNeedVerification(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))
	action := g.Query("action")

	// fmt.Println(limit, page)
	if action == "export-excel" {
		// connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		// preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_EXPORT_EXCEL_INVENTARIS)
		// if err != nil {
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

		// payload := []string{
		// 	"sfdsf", "sfds",
		// }

		// err = preQueueWorkerExcel.Publish(payload...)
		// if err != nil {
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

	} else {
		inventaris, totalFiltered, total, err := usecase.
			NewInventarisUseCase().
			SetDB(kernel.Kernel.Config.DB.Connection).
			SetRedisCache(kernel.Kernel.Config.REDIS.Cache).
			GetInventarisNeedVerificator(
				length,
				start,
				g,
			)

		// inventaris, totalFiltered, total, err := usecase.
		// 	NewInventarisUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.ELASTIC.Client).
		// 	GetFromElastic(
		// 		length,
		// 		start,
		// 		g,
		// 	)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}
		g.JSON(200, tools.HttpResponse{
			Message:         "success get data",
			Data:            inventaris,
			RecordsFiltered: totalFiltered,
			RecordsTotal:    total,
		})
	}

	return
}

func (a *InvoiceImpl) GetInventarisPemeliharaan(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))
	action := g.Query("action")

	// t, _ := g.Get("token_info")

	// fmt.Println(limit, page)
	if action == "export-excel" {
		// connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		// preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_EXPORT_EXCEL_INVENTARIS)
		// if err != nil {
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

		// payload := []string{
		// 	"sfdsf", "sfds",
		// }

		// err = preQueueWorkerExcel.Publish(payload...)
		// if err != nil {
		// 	g.JSON(400, err.Error())
		// 	g.Abort()
		// 	return
		// }

	} else {
		inventaris, totalFiltered, total, err := usecase.
			NewInventarisUseCase().
			SetDB(kernel.Kernel.Config.DB.Connection).
			SetRedisCache(kernel.Kernel.Config.REDIS.Cache).
			GetPemeliharaanInventaris(
				length,
				start,

				g,
			)

		// inventaris, totalFiltered, total, err := usecase.
		// 	NewInventarisUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.ELASTIC.Client).
		// 	GetFromElastic(
		// 		length,
		// 		start,
		// 		g,
		// 	)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}
		g.JSON(200, tools.HttpResponse{
			Message:         "success get data",
			Data:            inventaris,
			RecordsFiltered: totalFiltered,
			RecordsTotal:    total,
		})
	}

	return
}
