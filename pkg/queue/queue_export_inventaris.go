package queue

import (
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/consumer"
	"time"

	"github.com/adjust/rmq/v5"
)

type QueueExportInventaris struct{}

const QUEUE_EXPORT_EXCEL_INVENTARIS = "excel-worker"

func (q *QueueExportInventaris) Register(connection rmq.Connection) {

	taskExcelQueue, err := connection.OpenQueue(QUEUE_EXPORT_EXCEL_INVENTARIS)

	err = taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		panic(err)
	}
	_, err = taskExcelQueue.AddConsumer("task_export_inventaris", &consumer.TaskExportInventaris{
		DB:    kernel.Kernel.Config.DB.Connection,
		Redis: kernel.Kernel.Config.REDIS.RedisCache,
	})
	if err != nil {
		panic(err)
	}
}
