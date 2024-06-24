package kernel

import (
	"simadaservices/pkg/models"
)

type Config struct {
	SIMADA_SV_PORT_GT_WA string
	JwtKey               string
}

type core struct {
	Config       Config
	UserLoggedIn models.User
}

var Kernel *core

func NewKernel() *core {
	return &core{Config: Config{}}
}
