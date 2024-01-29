package rest

import (
	"simadaservices/cmd/service-transaction/kernel"
	"simadaservices/pkg/queue"
	"simadaservices/pkg/tools"
	"simadaservices/pkg/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvoiceApi interface {
	Get(gin *gin.Context)
}

type InvoiceImpl struct {
}

func NewInvoiceApi() InvoiceApi {
	return &InvoiceImpl{}
}

func (a *InvoiceImpl) Get(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))
	action := g.Query("action")

	// t, _ := g.Get("token_info")

	// fmt.Println(limit, page)
	if action == "export-excel" {
		connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_IMPORT_EXCEL_INVENTARIS)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		err = preQueueWorkerExcel.Publish("test")
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

	} else {
		// users, totalFiltered, total, err := usecase.
		// 	NewInventarisUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.ELASTIC.Client).
		// 	Get(
		// 		length,
		// 		start,
		// 		usecase.NewAuthUseCase(kernel.Kernel.Config.DB.Connection).
		// 			IsUserHasAccess(t.(jwt.MapClaims)["id"].(float64),
		// 				[]string{"transaksi.inventaris.delete"}),
		// 		g,
		// 	)

		inventaris, totalFiltered, total, err := usecase.
			NewInventarisUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.ELASTIC.Client).
			GetFromElastic(
				length,
				start,
				g,
			)
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
