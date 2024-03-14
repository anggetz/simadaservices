package queue

import (
	"service-worker/consumer"
	"time"

	"github.com/adjust/rmq/v5"
	"github.com/go-redis/cache/v9"
	"gorm.io/gorm"
)

type QueueExportRekapitulasi struct {
	DB    *gorm.DB
	Redis *cache.Cache
}

const QUEUE_EXPORT_EXCEL_REKAPITULASI = "rekapitulasi-worker"

func (q *QueueExportRekapitulasi) Register(connection rmq.Connection) {

	taskExcelQueue, err := connection.OpenQueue(QUEUE_EXPORT_EXCEL_REKAPITULASI)

	err = taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		panic(err)
	}
	_, err = taskExcelQueue.AddConsumer("task_export_rekapitulasi", &consumer.TaskExportRekapitulasi{
		DB:    q.DB,
		Redis: q.Redis,
	})

	if err != nil {
		panic(err)
	}
}
