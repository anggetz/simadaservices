package main

import (
	"encoding/json"
	"fmt"
	"libcore/pipelines"
	"libcore/tools"
	"log"
	"os"
	"service-pipeline/kernel"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"

	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

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

func setupRediCache() {
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

		setUpDB()
		setupElasticDB()
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
	setupElasticDB()
	setupRediCache()

	defer func() {
		db, _ := kernel.Kernel.Config.DB.Connection.DB()
		db.Close()
	}()

	// gocron.Every(10).Minutes().Lock().Do(pipelines.NewSyncInventaris(kernel.Kernel.Config.ELASTIC.Client, *kernel.Kernel.Config.DB.Connection).SyncPgToElastic)
	gocron.Every(10).Minutes().Do(pipelines.NewSyncInventaris().SetDB(kernel.Kernel.Config.DB.Connection).SetRedisCache(kernel.Kernel.Config.REDIS.Cache).CountInventaris)
	// gocron.Every(6).Hour().Do(pipelines.NewSyncInventaris(kernel.Kernel.Config.ELASTIC.Client, *kernel.Kernel.Config.DB.Connection).SyncPgToElastic)

	gocron.RunAll()
	<-gocron.Start()
}
