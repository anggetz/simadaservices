package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"simadaservices/cmd/service-gateway-whatsapp/kernel"
	"simadaservices/pkg/middlewares"
	"simadaservices/pkg/models"
	"text/template"
	"time"

	"github.com/anggetz/golangwa/pubsup"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	cron "github.com/robfig/cron/v3"
)

type Task struct{}

var errChan chan error

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

	r := gin.Default()

	// register router
	r.Use(middlewares.NewMiddlewareAuth(nc).SetJwtKey(kernel.Kernel.Config.JwtKey).TokenValidate)
	apiGroup := r.Group("/v1/wa")
	{
		apiGroup.GET("/login", func(ctx *gin.Context) {
			respondStruct := struct {
				Base64QR string
			}{}

			respond, err := nc.Request("simada_wa.login", nil, time.Second*30)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			err = json.Unmarshal(respond.Data, &respondStruct)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"data": respondStruct,
			})
		})

		apiGroup.GET("/try-template-get-model", func(ctx *gin.Context) {
			templateModel := ctx.Query("template-model")

			switch templateModel {
			case "pemanfaatan":
				{
					timeNow := time.Now()
					tglAkhir := timeNow.Add(24 * time.Minute)
					model := models.Pemanfaatan{
						ID:               0,
						PIDInventaris:    0,
						Peruntukan:       "FAKER",
						Umur:             20,
						UmurSatuan:       "Tahun",
						NoPerjanjian:     "XX/FAKER/00-00",
						TglMulai:         &timeNow,
						TglAkhir:         &tglAkhir,
						Mitra:            0,
						TipeKontribusi:   "Kontribusi",
						JumlahKontribusi: "",
						Aktif:            "1",
						Pegawai:          0,
						BagiHasil:        10000,
						Tetap:            10000,
						Draft:            "1",
					}

					ctx.JSON(http.StatusOK, gin.H{
						"data": model,
					})
				}
			default:
				{
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"message": "model not supported",
					})
					return
				}
			}
		})

		apiGroup.POST("/send-message", func(ctx *gin.Context) {
			respondStruct := struct {
				Ok bool
			}{}

			webReq := struct {
				Message string
				Jid     string
			}{}

			err := ctx.BindJSON(&webReq)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			timeNow := time.Now()
			tglAkhir := timeNow.Add(24 * time.Minute)
			model := models.Pemanfaatan{
				ID:               0,
				PIDInventaris:    0,
				Peruntukan:       "FAKER",
				Umur:             20,
				UmurSatuan:       "Tahun",
				NoPerjanjian:     "XX/FAKER/00-00",
				TglMulai:         &timeNow,
				TglAkhir:         &tglAkhir,
				Mitra:            0,
				TipeKontribusi:   "Kontribusi",
				JumlahKontribusi: "",
				Aktif:            "1",
				Pegawai:          0,
				BagiHasil:        10000,
				Tetap:            10000,
				Draft:            "1",
			}

			tmpl, err := template.New("test").Parse(webReq.Message)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			var tpl bytes.Buffer
			err = tmpl.Execute(&tpl, model)

			webReq.Message = tpl.String()
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			byWebReq, err := json.Marshal(webReq)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			respond, err := nc.Request("simada_wa.send", byWebReq, time.Second*10)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			err = json.Unmarshal(respond.Data, &respondStruct)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"data": respondStruct,
			})
		})

		apiGroup.GET("/get-devices", func(ctx *gin.Context) {

			respond, err := nc.Request("simada_wa.devices", nil, time.Second*10)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{
				"data": respond.Data,
			})
		})

		apiGroup.GET("/check-login", func(ctx *gin.Context) {

			isLoggedInResp := pubsup.IsLoggedInResponse{}

			respond, err := nc.Request("simada_wa.check-login", nil, time.Second*10)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			err = json.Unmarshal(respond.Data, &isLoggedInResp)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"data": isLoggedInResp,
			})
		})

		apiGroup.POST("/get-pair-code", func(ctx *gin.Context) {

			payload := pubsup.PairCode{}

			err := ctx.BindJSON(&payload)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			byPayload, err := json.Marshal(payload)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			respond, err := nc.Request("simada_wa.get-pair-code", byPayload, time.Second*10)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			getPairCodeResp := pubsup.PairCodeResponse{}

			err = json.Unmarshal(respond.Data, &getPairCodeResp)

			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"data": getPairCodeResp,
			})
		})

	}

	r.Run(":" + kernel.Kernel.Config.SIMADA_SV_PORT_GT_WA)
}
