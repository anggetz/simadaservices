package queue

import (
	"libcore/consumer"
	"log"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/go-redis/cache/v9"
	"gorm.io/gorm"
)

type QueueExportMutasiBMD struct {
	DB    *gorm.DB
	Redis *cache.Cache
}

const QUEUE_EXPORT_EXCEL_MUTASIBMD = "mutasibmd-worker"

func (q *QueueExportMutasiBMD) Register(connection rmq.Connection) {

	taskExcelQueue, _ := connection.OpenQueue(QUEUE_EXPORT_EXCEL_MUTASIBMD)
	err := taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		log.Println(err.Error())
	}
	_, err = taskExcelQueue.AddConsumer("task_export_mutasibmd", &consumer.TaskExportMutasiBMD{
		DB:    q.DB,
		Redis: q.Redis,
	})
	if err != nil {
		log.Println(err.Error())
	}
}
