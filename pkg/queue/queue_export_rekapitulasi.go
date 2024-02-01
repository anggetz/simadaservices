package queue

import (
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/consumer"
	"time"

	"github.com/adjust/rmq/v5"
)

type QueueExportRekapitulasi struct{}

const QUEUE_EXPORT_EXCEL_REKAPITULASI = "rekapitulasi-worker"

func (q *QueueExportRekapitulasi) Register(connection rmq.Connection) {

	taskExcelQueue, err := connection.OpenQueue(QUEUE_EXPORT_EXCEL_REKAPITULASI)

	err = taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		panic(err)
	}
	_, err = taskExcelQueue.AddConsumer("task_export_rekapitulasi", &consumer.TaskExportRekapitulasi{
		DB: kernel.Kernel.Config.DB.Connection,
	})
	if err != nil {
		panic(err)
	}
}
