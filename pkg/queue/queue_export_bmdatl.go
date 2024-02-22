package queue

import (
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/consumer"
	"time"

	"github.com/adjust/rmq/v5"
)

type QueueExportBMDATL struct{}

const QUEUE_EXPORT_EXCEL_BMDATL = "bmdatl-worker"

func (q *QueueExportBMDATL) Register(connection rmq.Connection) {

	taskExcelQueue, err := connection.OpenQueue(QUEUE_EXPORT_EXCEL_BMDATL)

	err = taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		panic(err)
	}
	_, err = taskExcelQueue.AddConsumer("task_export_bmdatl", &consumer.TaskExportBMDATL{
		DB:    kernel.Kernel.Config.DB.Connection,
		Redis: kernel.Kernel.Config.REDIS.RedisCache,
	})
	if err != nil {
		panic(err)
	}
}
