package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"simadaservices/cmd/service-worker/kernel"
	"simadaservices/pkg/models"
	usecase "simadaservices/pkg/usecase/report"
	"strconv"
	"strings"

	"github.com/adjust/rmq/v5"
	"github.com/go-redis/cache/v9"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TaskExportRekapitulasi struct {
	rmq.Consumer
	Payload TaskExportRekapitulasiPayload
	DB      *gorm.DB
	Redis   *cache.Cache
}

type TaskExportRekapitulasiPayload struct {
	Headers []string
	Data    []byte
}

const (
	REKAPITULASI_EXCEL_FILE_FOLDER = "rekapitulasi"
	REKAPITULASI_FORMAT_FILE_TIME  = "02-01-2006 15:04:05"
)

type QueryParamsRekapitulasi struct {
	Action                 string `json:"action"`
	F_Bulan                string `json:"f_bulan"`
	F_Jenis                string `json:"f_jenis"`
	F_Periode              string `json:"f_periode"`
	F_Tahun                string `json:"f_tahun"`
	Firstload              string `json:"firstload"`
	F_Penggunafilter       string `json:"f_penggunafilter"`
	F_Kuasapengguna_Filter string `json:"f_kuasapengguna_filter"`
	F_Subkuasa_Filter      string `json:"f_subkuasa_filter"`
	Penggunafilter         string `json:"penggunafilter"`
	Kuasapengguna_Filter   string `json:"kuasapengguna_filter"`
	Subkuasa_Filter        string `json:"subkuasa_filter"`
	Draw                   string `json:"draw"`
	F_Jenisrekap           string `json:"f_jenisrekap"`
	F_Jenisperiode         string `json:"f_jenisperiode"`
	QueueId                int    `json:"queue_id"`
}

func (t *TaskExportRekapitulasi) Consume(d rmq.Delivery) {
	var err error

	fmt.Println("performing task report rekapitulasi")

	// get params
	var params QueryParamsRekapitulasi
	err = json.Unmarshal([]byte(d.Payload()), &params)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return
	}

	opdname := usecase.OpdName{}
	opdname = usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetOpdName(d.Payload())

	timestr := t.DB.NowFunc().Format(REKAPITULASI_FORMAT_FILE_TIME)
	folderPath := os.Getenv("FOLDER_REPORT")
	folderReport := REKAPITULASI_EXCEL_FILE_FOLDER
	fileName := ""
	if opdname.Pengguna != "" {
		fileName = fileName + strings.ReplaceAll(opdname.Pengguna, " ", "_")
	} else {
		fileName = "Rekapitulasi"
	}
	if opdname.KuasaPengguna != "" {
		fileName = fileName + "|" + strings.ReplaceAll(opdname.KuasaPengguna, " ", "_")
	}
	if opdname.SubKuasaPengguna != "" {
		fileName = fileName + "|" + strings.ReplaceAll(opdname.SubKuasaPengguna, " ", "_")
	}
	fileName = fileName + "_" + timestr

	defer func(errors error) {
		if errors != nil {
			// error
			log.Println("Error:", errors.Error())
		} else {
			// success, update sukses status task queue
			tq := models.TaskQueue{}
			t.DB.First(&tq, "id = ?", params.QueueId)
			tq.Status = "success"
			if tq.TaskName == "" {
				tq.TaskName = "worker-export-" + REKAPITULASI_EXCEL_FILE_FOLDER
			}
			if tq.TaskType == "" {
				tq.TaskType = "export_report"
			}
			if tq.TaskUUID == "" {
				tq.TaskUUID = uuid.NewString()
			}
			tq.CallbackLink = fmt.Sprintf("%s/%s/%s", folderPath, folderReport, fileName)
			tq.UpdatedAt = t.DB.NowFunc()
			if err := t.DB.Save(&tq).Error; err != nil {
				log.Println("failed to update task")
			}
		}
	}(err)

	startTime := t.DB.NowFunc()
	log.Println("->> START EXPORT : ", opdname.Pengguna, "|", opdname.KuasaPengguna, "|", opdname.SubKuasaPengguna, " : ", startTime.String())
	// get data
	report, _, _, _, _, _ := usecase.NewReportRekapitulasiUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.REDIS.RedisCache).Export(0, 0, params.F_Periode, params.F_Penggunafilter,
		params.Penggunafilter, params.F_Kuasapengguna_Filter, params.Kuasapengguna_Filter, params.F_Subkuasa_Filter, params.Subkuasa_Filter,
		params.F_Tahun, params.F_Bulan, params.F_Jenis, params.Action, params.Firstload, params.Draw, params.F_Jenisrekap, params.F_Jenisperiode)
	log.Println(" -->> RES DATA : ", t.DB.NowFunc().String())

	log.Println(" -->> CREATE FILE : ", t.DB.NowFunc().String())
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()
	// Set the sheet name
	sheetName := "Sheet1"

	// Set header data
	headers := []string{"No", "Kode Barang", "Nama Barang", "Jumlah", "Nilai Perolehan (Rp)", "Nilai Atribusi (Rp)", "Nilai Perolehan Setelah Atribusi (Rp)",
		"Akumulasi Penyusutan (Rp)", "Nilai Buku (Rp)"}

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
		// fmt.Println("<<- INSERT DATA NO - ", no)

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
		f.SetCellValue(sheetName, "B"+newno, drow.KodeBarang)
		f.SetCellValue(sheetName, "C"+newno, drow.NamaBarang)
		f.SetCellValue(sheetName, "D"+newno, drow.Jumlah)
		f.SetCellValue(sheetName, "E"+newno, drow.NilaiPerolehan)
		f.SetCellValue(sheetName, "F"+newno, drow.NilaiAtribusi)
		f.SetCellValue(sheetName, "G"+newno, drow.NilaiPerolehanSetelahAtribusi)
		f.SetCellValue(sheetName, "H"+newno, drow.AkumulasiPenyusutan)
		f.SetCellValue(sheetName, "I"+newno, drow.NilaiBuku)

		no++
		totalRows++
	}
	log.Println(" -->> END INSERT DATA : ", t.DB.NowFunc().String())

	log.Println(" -->> START SAVE DATA : ", t.DB.NowFunc().String())
	os.MkdirAll(folderPath+"/"+folderReport, os.ModePerm)
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
