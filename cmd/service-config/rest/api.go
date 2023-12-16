package rest

import (
	"github.com/gin-gonic/gin"
)

type Api interface {
	Get(gin *gin.Context)
}

type ApiImpl struct {
}

func NewApi() Api {
	return &ApiImpl{}
}

func (a *ApiImpl) Get(g *gin.Context) {

}
