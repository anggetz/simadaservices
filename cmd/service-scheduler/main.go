package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"simadaservices/cmd/service-scheduler/kernel"
	"simadaservices/pkg/tools"
	schedulerLogic "simadaservices/pkg/usecase/scheduler"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/robfig/cron/v3"
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

	scheduler := cron.New(cron.WithLocation(time.Local))
	// stop scheduler tepat sebelum fungsi berakhir
	defer scheduler.Stop()

	scheduler.AddFunc("* * * * *", func() {
		// log.Println(">>> service worker : export bmd atl scheduler")
		// rest.NewApi().GetBmdAtl(kernel.Kernel.Config.DB.Connection, connectionRedis)
		log.Println("INFO", "Creating alert pemanfaatan")
		err := schedulerLogic.NewPemanfaatan(kernel.Kernel.Config.DB.Connection).Execute()
		if err != nil {
			panic(err)
		}
	}) // SETIAP HARI PUKUL 9 malam setiap hari

	go scheduler.Start()

	// _, err = schedulerLogic.NewSaldoAwal(kernel.Kernel.Config.DB.Connection).Execute(1, 2024)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// try the sc
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		fmt.Println(">>> error : hard exit on second signal (in case shutdown gets stuck")
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()

}
