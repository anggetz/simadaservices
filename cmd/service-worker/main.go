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
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

type Task struct{}

var errChan chan error

func setUpRedis() {

	go tools.LogErrors(errChan)
	connection, err := rmq.OpenConnection("consumer", "tcp", fmt.Sprintf("%s:%s", kernel.Kernel.Config.REDIS.Host, kernel.Kernel.Config.REDIS.Port), 1, errChan)
	if err != nil {
		fmt.Errorf("error", err.Error())
	} else {
		kernel.Kernel.Config.REDIS.Connection = &connection
	}

}

func main() {
	errChan = make(chan error, 10)
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

	setUpRedis()

	connectionRedis := *kernel.Kernel.Config.REDIS.Connection

	new(queue.QueueImportInventaris).Register(connectionRedis)
	fmt.Println("service worker already running")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
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
