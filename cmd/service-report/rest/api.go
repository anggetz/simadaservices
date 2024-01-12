package rest

import (
	"simadaservices/cmd/service-report/kernel"
	"simadaservices/pkg/tools"
	"simadaservices/pkg/usecase"
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

	data, recordsTotal, recordsFiltered, draw, summary_perpage, summary_page, err := usecase.NewReportUseCase(kernel.Kernel.Config.DB.Connection).GetInventaris(start, length, g)
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
	return
}

func (a *Api) GetRekapitulasi(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	data, recordsTotal, recordsFiltered, draw, summary_perpage, summary_page, err := usecase.NewReportUseCase(kernel.Kernel.Config.DB.Connection).GetRekapitulasi(start, length, g)
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
	return
}
