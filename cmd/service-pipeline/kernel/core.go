package kernel

import (
	"simadaservices/pkg/models"

	"github.com/adjust/rmq/v5"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/cache/v9"
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
	Cache      *cache.Cache
}

type elasticConfig struct {
	Address string
	Client  *elasticsearch.Client
}

type Config struct {
	SIMADA_SV_PORT_TRANSACTION string
	DB                         dbConfig
	REDIS                      redisConfig
	ELASTIC                    elasticConfig
}

type core struct {
	Config       Config
	UserLoggedIn models.User
}

var Kernel *core

func NewKernel() *core {
	return &core{Config: Config{}}
}
