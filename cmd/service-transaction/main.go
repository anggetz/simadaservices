package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simadaservices/cmd/service-transaction/kernel"
	"simadaservices/cmd/service-transaction/rest"
	"simadaservices/pkg/middlewares"
	"simadaservices/pkg/tools"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

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

func setupElasticDB() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			kernel.Kernel.Config.ELASTIC.Address,
		},
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("error", err.Error())
		return
	}

	kernel.Kernel.Config.ELASTIC.Client = es
}

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

		setUpDB()
		setUpRedis()
		// setupElasticDB()
		setupRediCache()

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
	setUpRedis()
	// setupElasticDB()
	setupRediCache()

	db, _ := kernel.Kernel.Config.DB.Connection.DB()
	defer db.Close()

	r := gin.Default()

	// register router
	apiGroup := r.Group("/v1/transaction")
	{
		apiGroup.Use(middlewares.NewMiddlewareAuth(nc).SetJwtKey(kernel.Kernel.Config.JwtKey).TokenValidate)
		apiGroupTransaction := apiGroup.Group("/inventaris")
		{
			apiGroupTransaction.GET("/get", rest.NewInvoiceApi().Get)
			apiGroupTransaction.GET("/get-queue", rest.NewInvoiceApi().GetQueueExportInventaris)
			apiGroupTransaction.GET("/get-inventaris-need-verification", rest.NewInvoiceApi().GetInventarisNeedVerification)
			apiGroupTransaction.GET("/get-inventaris-pemeliharaan", rest.NewInvoiceApi().GetInventarisPemeliharaan)
		}

		apiGroupHome := apiGroup.Group("/home")
		{
			apiGroupHome.GET("/get-total-aset", rest.NewHomeApi().GetTotalAset)
			apiGroupHome.GET("/get-nilai-aset", rest.NewHomeApi().GetNilaiAsset)
			apiGroupHome.GET("/get-nilai-aset-by-kode-jenis", rest.NewHomeApi().GetNilaiAssetByKodeJenis)
		}
	}

	r.Run(":" + kernel.Kernel.Config.SIMADA_SV_PORT_TRANSACTION)

	nc.Drain()
}
