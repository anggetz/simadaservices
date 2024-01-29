package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simadaservices/cmd/service-report/kernel"
	"simadaservices/cmd/service-report/rest"
	"simadaservices/pkg/middlewares"
	"simadaservices/pkg/tools"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func setUpDB() {

	if kernel.Kernel.Config.DB.Connection != nil {
		sqlDb, err := kernel.Kernel.Config.DB.Connection.DB()
		if err != nil {
			panic(err)
		}

		err = sqlDb.Close()
		if err != nil {
			panic(err)
		}
		fmt.Println("close the database and re-creating new one")
	}

	fmt.Println("setting up database", kernel.Kernel.Config.DB)

	db, err := tools.NewDatabase().GetGormConnection(
		kernel.Kernel.Config.DB.Host,
		kernel.Kernel.Config.DB.Port,
		kernel.Kernel.Config.DB.User,
		kernel.Kernel.Config.DB.Password,
		kernel.Kernel.Config.DB.Database,
		kernel.Kernel.Config.DB.TimeZone)

	if err != nil {
		panic(err)
	}

	kernel.Kernel.Config.DB.Connection = db
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	kernel.Kernel = kernel.NewKernel()
	// register nats
	// Connect to a server
	nc, _ := nats.Connect(fmt.Sprintf("%s:%s", os.Getenv("NATS_HOST"), os.Getenv("NATS_PORT")))
	nc.Subscribe("config.share", func(msg *nats.Msg) {
		err := json.Unmarshal(msg.Data, &kernel.Kernel.Config)
		if err != nil {
			panic(err)
		}

		log.Println("new config receive", kernel.Kernel.Config)
	})

	msg, err := nc.Request("config.get", []byte(""), time.Second*10)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(msg.Data, &kernel.Kernel.Config)
	if err != nil {
		panic(err)
	}

	log.Println("config receive", kernel.Kernel.Config)

	setUpDB()
	db, _ := kernel.Kernel.Config.DB.Connection.DB()
	defer db.Close()

	r := gin.Default()

	// register router
	apiGroup := r.Group("/v1/report").Use(middlewares.NewMiddlewareAuth(nc).TokenValidate)
	apiGroup.GET("/get-inventaris", rest.NewApi().GetInventaris)
	apiGroup.GET("/get-rekapitulasi", rest.NewApi().GetRekapitulasi)
	apiGroup.GET("/get-total-rekapitulasi", rest.NewApi().GetTotalRekapitulasi)
	apiGroup.GET("/get-bmdatl", rest.NewApi().GetBmdAtl)
	apiGroup.GET("/get-bmdatl-totalrecords", rest.NewApi().GetBmdAtlTotalRecords)
	apiGroup.GET("/get-total-bmdatl", rest.NewApi().GetTotalBmdAtl)
	// apiGroup.GET("/get", rest.NewApi().Get)

	r.Run(":" + kernel.Kernel.Config.SIMADA_SV_PORT_REPORT)

	nc.Drain()
}
