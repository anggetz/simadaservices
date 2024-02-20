package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/models"
	usecase "simadaservices/pkg/usecase"
	usecase2 "simadaservices/pkg/usecase/report"
	"strconv"

	"github.com/adjust/rmq/v5"
	"github.com/go-redis/cache/v9"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TaskExportInventaris struct {
	rmq.Consumer
	Payload TaskExportInventarisPayload
	DB      *gorm.DB
	Redis   *cache.Cache
}

type TaskExportInventarisPayload struct {
	Headers []string
	Data    []interface{}
}

const (
	INVEN_EXCEL_FILE_FOLDER = "inventaris"
	INVEN_FORMAT_FILE_TIME  = "02-01-2006 15:04:05"
)

func (t *TaskExportInventaris) Consume(d rmq.Delivery) {
	var err error
	// insert data

	// get params
	var params usecase.QueryParamInventaris
	err = json.Unmarshal([]byte(d.Payload()), &params)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err.Error())
		return
	}

	opdname := usecase2.OpdName{}
	opdname = usecase2.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetOpdName(d.Payload())

	folderPath := os.Getenv("FOLDER_REPORT")
	folderReport := INVEN_EXCEL_FILE_FOLDER
	os.MkdirAll(folderPath+"/"+folderReport, os.ModePerm)
	timestr := t.DB.NowFunc().Format(INVEN_FORMAT_FILE_TIME)
	fileName := opdname.Pengguna + ":" + opdname.KuasaPengguna + ":" + opdname.SubKuasaPengguna + "-" + params.TokenUsername + "_" + timestr

	defer func(errors error) {
		if errors != nil {
			// error
			log.Println("Error:", errors.Error())
		} else {
			// success, update sukses status task queue
			tq := models.TaskQueue{}
			t.DB.First(&tq, "id = ?", params.QueueId)
			tq.Status = "success"
			tq.CallbackLink = fmt.Sprintf("%s/%s/%s", folderPath, folderReport, fileName)
			tq.UpdatedAt = t.DB.NowFunc()
			if err := t.DB.Save(&tq).Error; err != nil {
				log.Println("failed to update task")
			}
		}
	}(err)

	fmt.Println("performing task report inventaris")

	startTime := t.DB.NowFunc()
	log.Println("->> START EXPORT : ", opdname.Pengguna, "|", opdname.KuasaPengguna, "|", opdname.SubKuasaPengguna, " : ", startTime.String())
	// get data
	report, err := usecase.NewInventarisUseCase().
		SetDB(kernel.Kernel.Config.DB.Connection).
		SetRedisCache(kernel.Kernel.Config.REDIS.RedisCache).
		GetExportInventaris(params)

	if err != nil {
		log.Println("Error get data: ", err.Error())
		return
	}

	log.Println(" -->> RES DATA : ", t.DB.NowFunc().String())
	log.Println(" -->> CREATE FILE : ", t.DB.NowFunc().String())

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("failed to close file", err.Error())
		}
	}()
	// Set the sheet name
	sheetName := "Sheet1"

	// Set header data
	headers := []string{"No", "ID Publish", "Kode Barang", "Nomor Register", "Nama Barang", "Cara Perolehan",
		"Tahun Perolehan", "Kondisi", "Pengguna Barang", "Harga Satuan", "Status Verifikasi"}

	// Create a bold style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	log.Println(" -->> START INSERT DATA : ", t.DB.NowFunc().String())
	// Set the header row and make it bold
	cellName := ""
	addData := 1
	for colIndex, header := range headers {
		cellName, _ = excelize.CoordinatesToCellName(colIndex+addData, 1)
		f.SetCellValue(sheetName, cellName, header)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
	}

	// Set value of a cell.
	no := 1
	noSheet := 1
	totalRows := 1
	newno := ""
	for _, drow := range report {
		// if sheet not enough
		if totalRows > 1048570 {
			noSheet++
			sheetName = "Sheet" + strconv.Itoa(noSheet)
			indexSheet, _ := f.NewSheet(sheetName)
			f.SetActiveSheet(indexSheet)
			totalRows = 1

			// Set the header row and make it bold
			cellName := ""
			addData := 1
			for colIndex, header := range headers {
				cellName, _ = excelize.CoordinatesToCellName(colIndex+addData, 1)
				f.SetCellValue(sheetName, cellName, header)
				f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
			}
		}

		newno = strconv.Itoa(totalRows + 1)
		f.SetCellValue(sheetName, "A"+newno, no)
		f.SetCellValue(sheetName, "B"+newno, drow.IdPublish)
		f.SetCellValue(sheetName, "C"+newno, drow.KodeBarang)
		f.SetCellValue(sheetName, "D"+newno, drow.Noreg)
		f.SetCellValue(sheetName, "E"+newno, drow.NamaRekAset)
		f.SetCellValue(sheetName, "F"+newno, drow.Perolehan)
		f.SetCellValue(sheetName, "G"+newno, drow.TahunPerolehan)
		f.SetCellValue(sheetName, "H"+newno, drow.Kondisi)
		f.SetCellValue(sheetName, "I"+newno, drow.PenggunaBarang)
		f.SetCellValue(sheetName, "J"+newno, drow.HargaSatuan)
		f.SetCellValue(sheetName, "K"+newno, drow.StatusVerifikasi)

		no++
		totalRows++
	}
	log.Println(" -->> END INSERT DATA : ", t.DB.NowFunc().String())

	log.Println(" -->> START SAVE DATA : ", t.DB.NowFunc().String())
	// fileName := " Export Inventaris " + "-" + strconv.Itoa(params.QueueId) + "-" + timestr
	if err := f.SaveAs(fmt.Sprintf("%s/%s/%s.xlsx", folderPath, folderReport, fileName)); err != nil {
		log.Println("ERROR", err.Error())
		err = d.Reject()
		if err != nil {
			log.Println("REJECT ERROR", err.Error())
		}
		return
	}
	endTime := t.DB.NowFunc()
	duration := endTime.Sub(startTime)
	log.Println(" -->> Duration : ", duration.String())
	log.Println("->> END EXPORT : ", opdname.Pengguna, "|", opdname.KuasaPengguna, "|", opdname.SubKuasaPengguna, " : ", startTime.String())

	err = d.Ack()
	if err != nil {
		log.Println("ACK ERROR!", err.Error())
	}

	fmt.Println("ending task")

}
