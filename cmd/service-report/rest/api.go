package rest

import (
	"simadaservices/pkg/tools"

	"github.com/gin-gonic/gin"
)

type Api interface {
	Get(*gin.Context)
}

type ApiImpl struct{}

func NewApi() Api {
	return &ApiImpl{}
}

func (a *ApiImpl) Get(g *gin.Context) {

	g.JSON(200, tools.HttpResponse{
		Message: "success get data",
	})
}
