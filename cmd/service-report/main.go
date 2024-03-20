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

	"github.com/adjust/rmq/v5"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
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

func setUpRedis() {
	errChan := make(chan error, 10)
	go tools.LogErrors(errChan)
	connection, err := rmq.OpenConnection("consumer", "tcp", fmt.Sprintf("%s:%s", kernel.Kernel.Config.REDIS.Host, kernel.Kernel.Config.REDIS.Port), 1, errChan)
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
		fmt.Println("setting up redis connection")
		kernel.Kernel.Config.REDIS.Connection = &connection
	}

}

func setupRediCache() {
	fmt.Println("connect to", fmt.Sprintf("%s:%s", kernel.Kernel.Config.REDIS.Host, kernel.Kernel.Config.REDIS.Port))

	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"server1": fmt.Sprintf("%s:%s", kernel.Kernel.Config.REDIS.Host, kernel.Kernel.Config.REDIS.Port),
		},
	})

	mycache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	kernel.Kernel.Config.REDIS.Cache = mycache
}

func main() {
	// Create or open a log file for writing
	currentTime := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile("storage/logs/"+currentTime+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer logFile.Close()

	// Set the log output to the log file
	log.SetOutput(logFile)

	err = godotenv.Load()
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

	setUpRedis()
	setUpDB()
	setupRediCache()

	db, _ := kernel.Kernel.Config.DB.Connection.DB()
	defer db.Close()

	r := gin.Default()

	// register router
	apiGroup := r.Group("/v1/report").Use(middlewares.NewMiddlewareAuth(nc).TokenValidate)

	// inventaris
	apiGroup.GET("/get-inventaris", rest.NewApi().GetInventaris)

	// rekapitulasi
	apiGroup.GET("/get-rekapitulasi", rest.NewApi().GetRekapitulasi)
	apiGroup.GET("/get-total-rekapitulasi", rest.NewApi().GetTotalRekapitulasi)
	apiGroup.GET("/export-rekapitulasi", rest.NewApi().ExportRekapitulasi)

	// rekapitulasi
	apiGroup.GET("/get-mutasibmd", rest.NewApi().GetMutasiBmd)
	apiGroup.GET("/get-total-mutasibmd", rest.NewApi().GetTotalMutasiBmd)
	apiGroup.GET("/export-mutasibmd", rest.NewApi().ExportMutasiBmd)

	// bmdatl
	apiGroup.GET("/get-bmdatl", rest.NewApi().GetBmdAtl)
	apiGroup.GET("/get-total-bmdatl", rest.NewApi().GetTotalBmdAtl)
	apiGroup.GET("/export-bmdatl", rest.NewApi().ExportBmdAtl)

	// bmdtanah
	apiGroup.GET("/get-bmdtanah", rest.NewApi().GetBmdTanah)
	apiGroup.GET("/get-total-bmdtanah", rest.NewApi().GetTotalBmdTanah)
	apiGroup.GET("/export-bmdtanah", rest.NewApi().ExportBmdTanah)

	// FILE EXPORT
	r.GET("/v1/report/download-file", rest.NewApi().GetFileExport)
	apiGroup.GET("/file-list", rest.NewApi().FileListExport)
	apiGroup.GET("/delete-file", rest.NewApi().DeleteFileExport)

	r.Run(":" + kernel.Kernel.Config.SIMADA_SV_PORT_REPORT)

	nc.Drain()
}
