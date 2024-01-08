package kernel

import "github.com/adjust/rmq/v5"

type dbConfig struct {
	Host     string
	User     string
	Port     string
	Password string
	Database string
	TimeZone string
}

type Config struct {
	SIMADA_SV_PORT_AUTH        string
	SIMADA_SV_PORT_TRANSACTION string
	SIMADA_SV_PORT_REPORT      string
	DB                         dbConfig
	REDIS                      redisConfig
}

type redisConfig struct {
	Host       string
	Port       string
	Connection *rmq.Connection
}

type core struct {
	Config Config
}

var Kernel *core

func NewKernel() *core {
	return &core{Config: Config{}}
}
