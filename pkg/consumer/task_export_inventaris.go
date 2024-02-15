package consumer

import (
	"fmt"

	"github.com/adjust/rmq/v5"
	"github.com/xuri/excelize/v2"
)

type TaskExportInventaris struct {
	rmq.Consumer
	Payload TaskExportInventarisPayload
}

type TaskExportInventarisPayload struct {
	Headers []string
	Data    []interface{}
}

func (t *TaskExportInventaris) Consume(d rmq.Delivery) {
	var err error
	// insert data

	defer func(errors error) {
		if errors != nil {
			// sucess
		} else {
			//  update status error
		}

	}(err)

	fmt.Println("performing task")

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create a new sheet.
	index, err := f.NewSheet("Sheet2")
	if err != nil {
		fmt.Println("ERROR", err.Error)
		err = d.Reject()
		if err != nil {
			fmt.Println("REJECT ERROR", err.Error)
		}
		return

	}
	// Set value of a cell.
	f.SetCellValue("Sheet2", "A2", "Hello world.")
	f.SetCellValue("Sheet1", "B2", 100)
	// Set active sheet of the workbook.
	f.SetActiveSheet(index)
	// Save spreadsheet by the given path.
	if err := f.SaveAs("Book1.xlsx"); err != nil {

		fmt.Println("ERROR", err.Error)
		err = d.Reject()
		if err != nil {
			fmt.Println("REJECT ERROR", err.Error)
		}
		return
	}

	err = d.Ack()
	if err != nil {
		fmt.Println("ACK ERROR!", err.Error())
	}

}
