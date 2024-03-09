package main

import (
	"encoding/json"
	"fmt"
	"libcore/models"
	"libcore/tools"
	"libcore/usecase"
	"log"
	"os"
	"service-auth/kernel"
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

		setUpDB()

		log.Println("new config receive", kernel.Kernel.Config)
	})

	nc.Subscribe("auth.validate", func(msg *nats.Msg) {
		token := struct {
			Token    string
			User     models.User
			Response bool
		}{}

		err := json.Unmarshal(msg.Data, &token)
		if err != nil {
			panic(err)
		}

		token.User, token.Response = usecase.NewAuthUseCase(kernel.Kernel.Config.DB.Connection).ValidateToken(token.Token)

		tokenMarshalled, err := json.Marshal(token)
		if err != nil {
			panic(err)
		}

		msg.Respond([]byte(tokenMarshalled))
	})

	msg, err := nc.Request("config.get", []byte(""), time.Second*10)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(msg.Data, &kernel.Kernel.Config)
	if err != nil {
		panic(err)
	}

	setUpDB()
	db, _ := kernel.Kernel.Config.DB.Connection.DB()
	defer db.Close()
	log.Println("config receive", kernel.Kernel.Config)

	r := gin.Default()

	// register router
	// apiGroup := r.Group("/v1/auth")
	// apiGroup.GET("/get")

	r.Run(":" + kernel.Kernel.Config.SIMADA_SV_PORT_AUTH)

	nc.Drain()
}
