package queue

import (
	"log"
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/consumer"
	"time"

	"github.com/adjust/rmq/v5"
)

type QueueExportMutasiBMD struct{}

const QUEUE_EXPORT_EXCEL_MUTASIBMD = "mutasibmd-worker"

func (q *QueueExportMutasiBMD) Register(connection rmq.Connection) {

	taskExcelQueue, _ := connection.OpenQueue(QUEUE_EXPORT_EXCEL_MUTASIBMD)
	err := taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		log.Println(err.Error())
	}
	_, err = taskExcelQueue.AddConsumer("task_export_mutasibmd", &consumer.TaskExportMutasiBMD{
		DB:    kernel.Kernel.Config.DB.Connection,
		Redis: kernel.Kernel.Config.REDIS.RedisCache,
	})
	if err != nil {
		log.Println(err.Error())
	}
}
