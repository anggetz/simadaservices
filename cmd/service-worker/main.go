package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/queue"
	"simadaservices/pkg/tools"
	"syscall"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/go-redis/cache/v9"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	cron "github.com/robfig/cron/v3"
)

type Task struct{}

var errChan chan error

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

	go tools.LogErrors(errChan)
	connection, err := rmq.OpenConnection("consumer", "tcp", fmt.Sprintf("%s:%s", kernel.Kernel.Config.REDIS.Host, kernel.Kernel.Config.REDIS.Port), 1, errChan)
	if err != nil {
		fmt.Println("error", err.Error())
	} else {
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

	kernel.Kernel.Config.REDIS.RedisCache = mycache
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

	errChan = make(chan error, 10)
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	kernel.Kernel = kernel.NewKernel()

	// set scheduler berdasarkan zona waktu sesuai kebutuhan
	// jakartaTime, err :=
	// if err != nil {
	// 	log.Fatal("Error loading Asia Jakarta")
	// }

	scheduler := cron.New(cron.WithLocation(time.Local))
	// stop scheduler tepat sebelum fungsi berakhir
	defer scheduler.Stop()

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
	setUpRedis()
	setupRediCache()

	db, _ := kernel.Kernel.Config.DB.Connection.DB()
	defer db.Close()
	connectionRedis := *kernel.Kernel.Config.REDIS.Connection

	new(queue.QueueExportInventaris).Register(connectionRedis)
	new(queue.QueueExportBMDATL).Register(connectionRedis)
	new(queue.QueueExportRekapitulasi).Register(connectionRedis)
	new(queue.QueueExportMutasiBMD).Register(connectionRedis)

	// set task yang akan dijalankan scheduler
	scheduler.AddFunc("00 21 * * *", func() {
		// log.Println(">>> service worker : export bmd atl scheduler")
		// rest.NewApi().GetBmdAtl(kernel.Kernel.Config.DB.Connection, connectionRedis)
	}) // SETIAP HARI PUKUL 9 malam setiap hari
	scheduler.AddFunc("01 0 * * *", func() {
		// log.Println(">>> service worker : reminder penggunaan sementara")
		// rest.NewApi().GetReminderPenggunaanSementara(kernel.Kernel.Config.DB.Connection, connectionRedis)
	}) // SETIAP HARI PUKUL 00 lebih 1 menit malam setiap hari
	go scheduler.Start()

	fmt.Println("service worker already running")
	fmt.Println("service cron already running")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		fmt.Println(">>> error : hard exit on second signal (in case shutdown gets stuck")
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

	// clean the open queue
	openQueues, err := connectionRedis.GetOpenQueues()
	if err != nil {
		panic(err)
	}

	for _, queue := range openQueues {
		taskQueue, _ := connectionRedis.OpenQueue(queue)
		if err != nil {
			fmt.Println(">>> error : ", err.Error())
			errChan <- err
		} else {
			readyCount, rejectedCount, err := taskQueue.Destroy()
			fmt.Println(queue)
			fmt.Println("destroying queues with ready count", readyCount)
			fmt.Println("destroying queues with reject count", rejectedCount)
			if err != nil {
				errChan <- err
			}
		}
	}

	<-connectionRedis.StopAllConsuming() // wait for all Consume() calls to finish
}
