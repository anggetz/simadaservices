package rest

import (
	"fmt"
	"simadaservices/cmd/service-transaction/kernel"
	"simadaservices/pkg/queue"
	"simadaservices/pkg/tools"
	"simadaservices/pkg/usecase"
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
		connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_IMPORT_EXCEL_INVENTARIS)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		payload := []string{
			"sfdsf", "sfds",
		}

		err = preQueueWorkerExcel.Publish(payload...)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		g.JSON(200, tools.HttpResponse{
			Message: "Please wait! file exporting.",
		})
		return
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
		connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_IMPORT_EXCEL_INVENTARIS)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		payload := []string{
			"sfdsf", "sfds",
		}

		err = preQueueWorkerExcel.Publish(payload...)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

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
		connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_IMPORT_EXCEL_INVENTARIS)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		payload := []string{
			"sfdsf", "sfds",
		}

		err = preQueueWorkerExcel.Publish(payload...)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

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
