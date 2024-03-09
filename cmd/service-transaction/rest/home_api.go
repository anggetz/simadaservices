package rest

import (
	"libcore/tools"
	"libcore/usecase"
	"service-transaction/kernel"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type HomeApi interface {
	GetTotalAset(*gin.Context)
	GetNilaiAsset(*gin.Context)
	GetNilaiAssetByKodeJenis(*gin.Context)
}

type homeApiImpl struct{}

func NewHomeApi() HomeApi {
	return &homeApiImpl{}
}

func (h *homeApiImpl) GetTotalAset(g *gin.Context) {

	t, _ := g.Get("token_info")

	totalAset, err := usecase.NewHomeUseCase(kernel.Kernel.Config.DB.Connection).GetTotalAset(t.(jwt.MapClaims))
	if err != nil {
		g.AbortWithError(400, err)
		return
	}

	g.JSON(200, tools.HttpResponse{
		Message: "Successfully get total data",
		Data:    totalAset,
	})
}

func (h *homeApiImpl) GetNilaiAsset(g *gin.Context) {

	t, _ := g.Get("token_info")

	nilaiAsset, err := usecase.NewHomeUseCase(kernel.Kernel.Config.DB.Connection).GetNilaiAsset(t.(jwt.MapClaims), g)
	if err != nil {
		g.AbortWithError(400, err)
		return
	}

	g.JSON(200, tools.HttpResponse{
		Message: "Successfully get nilai asset",
		Data:    nilaiAsset,
	})
}

func (h *homeApiImpl) GetNilaiAssetByKodeJenis(g *gin.Context) {

	t, _ := g.Get("token_info")

	nilaiAsset, err := usecase.NewHomeUseCase(kernel.Kernel.Config.DB.Connection).GetNilaiAssetGroupByKodeJenis(t.(jwt.MapClaims))
	if err != nil {
		g.AbortWithError(400, err)
		return
	}

	g.JSON(200, tools.HttpResponse{
		Message: "Successfully get nilai asset",
		Data:    nilaiAsset,
	})
}
