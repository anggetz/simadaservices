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
	limit, _ := strconv.Atoi(g.Query("limit"))
	page, _ := strconv.Atoi(g.Query("page"))

	// fmt.Println(limit, page)
	users, total, err := usecase.NewInventarisUseCase(kernel.Kernel.Config.DB.Connection).Get(limit, page, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponse{
		Message: "success get data",
		Data:    users,
		Total:   total,
	})
	return
}
