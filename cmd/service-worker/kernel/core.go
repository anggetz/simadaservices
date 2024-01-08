package kernel

import (
	"simadaservices/pkg/models"

	"github.com/adjust/rmq/v5"
	"gorm.io/gorm"
)

type dbConfig struct {
	Host       string
	User       string
	Port       string
	Password   string
	Database   string
	TimeZone   string
	Connection *gorm.DB
}

type redisConfig struct {
	Host       string
	Port       string
	Connection *rmq.Connection
}

type Config struct {
	SIMADA_SV_PORT_TRANSACTION string
	DB                         dbConfig
	REDIS                      redisConfig
}

type core struct {
	Config       Config
	UserLoggedIn models.User
}

var Kernel *core

func NewKernel() *core {
	return &core{Config: Config{}}
}
