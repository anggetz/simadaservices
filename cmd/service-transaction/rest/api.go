package rest

import (
	"simadaservices/cmd/service-transaction/kernel"
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

	// fmt.Println(limit, page)
	users, totalFiltered, total, err := usecase.NewInventarisUseCase(kernel.Kernel.Config.DB.Connection).Get(length, start, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponse{
		Message:         "success get data",
		Data:            users,
		RecordsFiltered: totalFiltered,
		RecordsTotal:    total,
	})
	return
}
