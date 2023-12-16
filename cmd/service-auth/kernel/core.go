package kernel

import "gorm.io/gorm"

type dbConfig struct {
	Host       string
	User       string
	Port       string
	Password   string
	Database   string
	TimeZone   string
	Connection *gorm.DB
}

type Config struct {
	SIMADA_SV_PORT_AUTH string
	DB                  dbConfig
}

type core struct {
	Config Config
}

var Kernel *core

func NewKernel() *core {
	return &core{Config: Config{}}
}
