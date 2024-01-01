package rest

import (
	"simadaservices/cmd/service-transaction/kernel"
	"simadaservices/pkg/tools"
	"simadaservices/pkg/usecase"
	"strconv"

	"github.com/dgrijalva/jwt-go"
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
	// action := g.Query("action")

	t, _ := g.Get("token_info")

	// fmt.Println(limit, page)
	users, totalFiltered, total, err := usecase.
		NewInventarisUseCase(kernel.Kernel.Config.DB.Connection).
		Get(
			length,
			start,
			usecase.NewAuthUseCase(kernel.Kernel.Config.DB.Connection).
				IsUserHasAccess(t.(jwt.MapClaims)["id"].(float64),
					[]string{"transaksi.inventaris.delete"}),
			g,
		)
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
