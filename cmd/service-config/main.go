package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simadaservices/cmd/service-config/kernel"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()

	// register router
	// apiGroup := r.Group("/v1/config")
	// apiGroup.GET("/get", rest.NewApi().Get)

	// register nats
	// Connect to a server
	nc, _ := nats.Connect(fmt.Sprintf("%s:%s", os.Getenv("NATS_HOST"), os.Getenv("NATS_PORT")))

	kernel.Kernel = kernel.NewKernel()
	kernel.Kernel.Config.SIMADA_SV_PORT_AUTH = os.Getenv("SIMADA_SV_PORT_AUTH")
	kernel.Kernel.Config.SIMADA_SV_PORT_TRANSACTION = os.Getenv("SIMADA_SV_PORT_TRANSACTION")
	kernel.Kernel.Config.SIMADA_SV_PORT_REPORT = os.Getenv("SIMADA_SV_PORT_REPORT")
	kernel.Kernel.Config.DB.Database = os.Getenv("SIMADA_DB_PG_DB")
	kernel.Kernel.Config.DB.Host = os.Getenv("SIMADA_DB_PG_HOST")
	kernel.Kernel.Config.DB.Port = os.Getenv("SIMADA_DB_PG_PORT")
	kernel.Kernel.Config.DB.User = os.Getenv("SIMADA_DB_PG_USER")
	kernel.Kernel.Config.DB.Password = os.Getenv("SIMADA_DB_PG_PASSWORD")
	kernel.Kernel.Config.DB.TimeZone = os.Getenv("SIMADA_DB_PG_TIMEZONE")

	kernel.Kernel.Config.REDIS.Host = os.Getenv("REDIS_HOST")
	kernel.Kernel.Config.REDIS.Port = os.Getenv("REDIS_PORT")

	kernel.Kernel.Config.JwtKey = os.Getenv("JWT_KEY")

	kernel.Kernel.Config.ELASTIC.Address = os.Getenv("ELASTIC_ADDRESS")

	confMarshalled, err := json.Marshal(kernel.Kernel.Config)
	if err != nil {
		panic(err)
	}
	nc.Subscribe("config.get", func(msg *nats.Msg) {
		msg.Respond(confMarshalled)
	})

	err = nc.Publish("config.share", confMarshalled)
	if err != nil {
		panic(err)
	}

	r.Run("localhost:8001")
}
