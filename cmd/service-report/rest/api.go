package rest

import (
	"simadaservices/cmd/service-report/kernel"
	"simadaservices/pkg/queue"
	"simadaservices/pkg/tools"
	usecase "simadaservices/pkg/usecase/report"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Api struct{}

func NewApi() *Api {
	return &Api{}
}

func (a *Api) GetInventaris(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))
	action := g.Query("action")

	if action == "export-excel" {
		connectionRedis := *kernel.Kernel.Config.REDIS.Connection
		preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_IMPORT_EXCEL_INVENTARIS)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		err = preQueueWorkerExcel.Publish("report-inventaris")
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

	} else {
		data, recordsTotal, recordsFiltered, draw, summary_perpage, summary_page, err := usecase.NewReportInventarisUseCase(kernel.Kernel.Config.DB.Connection).Get(start, length, g)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}

		g.JSON(200, tools.HttpResponseReport{
			Message:         "success get data",
			Data:            data,
			RecordsTotal:    recordsTotal,
			RecordsFiltered: recordsFiltered,
			Draw:            draw,
			SummaryPerPage:  summary_perpage,
			SummaryPage:     summary_page,
		})
	}
	return
}

func (a *Api) GetRekapitulasi(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	data, recordsTotal, recordsFiltered, draw, summary_perpage, err := usecase.NewReportRekapitulasiUseCase(kernel.Kernel.Config.DB.Connection).Get(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message:         "success get data",
		Data:            data,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Draw:            draw,
		SummaryPerPage:  summary_perpage,
	})
	return
}

func (a *Api) GetTotalRekapitulasi(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	summary_page, err := usecase.NewReportRekapitulasiUseCase(kernel.Kernel.Config.DB.Connection).GetTotal(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message: "success get data",
		Data:    summary_page,
	})
	return
}

func (a *Api) GetBmdAtl(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	data, recordsTotal, recordsFiltered, draw, summary_perpage, err := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection).Get(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message:         "success get data",
		Data:            data,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Draw:            draw,
		SummaryPerPage:  summary_perpage,
	})
	return
}

func (a *Api) GetBmdAtlTotalRecords(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	recordsTotal, err := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection).GetTotalRecords(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message: "success get data",
		Data:    recordsTotal,
	})
	return
}

func (a *Api) GetTotalBmdAtl(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	summary_page, err := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection).GetTotal(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message: "success get data",
		Data:    summary_page,
	})
	return
}
