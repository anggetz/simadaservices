package queue

import (
	"simadaservices/pkg/consumer"
	"time"

	"github.com/adjust/rmq/v5"
)

type QueueImportInventaris struct{}

const QUEUE_IMPORT_EXCEL_INVENTARIS = "excel-worker"

func (q *QueueImportInventaris) Register(connection rmq.Connection) {

	taskExcelQueue, err := connection.OpenQueue(QUEUE_IMPORT_EXCEL_INVENTARIS)

	err = taskExcelQueue.StartConsuming(10, 20*time.Second)
	if err != nil {
		panic(err)
	}
	_, err = taskExcelQueue.AddConsumer("task_import_inventaris", &consumer.TaskExportInventaris{})
	if err != nil {
		panic(err)
	}
}
