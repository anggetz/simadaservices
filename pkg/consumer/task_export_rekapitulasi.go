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

	"github.com/adjust/rmq/v5"
	"github.com/go-redis/cache/v9"
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
	REKAPITULASI_EXCEL_FILE_FOLDER = "Rekapitulasi"
	REKAPITULASI_FORMAT_FILE_TIME  = "02-01-2006 15:04:05"
)

type QueryParamsRekapitulasi struct {
	Action                 string `form:"action"`
	F_Bulan                string `form:"f_bulan"`
	F_Jenis                string `form:"f_jenis"`
	F_Periode              string `form:"f_periode"`
	F_Tahun                string `form:"f_tahun"`
	Firstload              string `form:"firstload"`
	F_Penggunafilter       string `form:"f_penggunafilter"`
	F_Kuasapengguna_Filter string `form:"f_kuasapengguna_filter"`
	F_Subkuasa_Filter      string `form:"f_subkuasa_filter"`
	Penggunafilter         string `form:"penggunafilter"`
	Kuasapengguna_Filter   string `form:"kuasapengguna_filter"`
	Subkuasa_Filter        string `form:"subkuasa_filter"`
	Draw                   string `form:"draw"`
	F_Jenisrekap           string `form:"f_jenisrekap"`
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

	startTime := t.DB.NowFunc()
	log.Println("->> START EXPORT : ", opdname.Pengguna, "|", opdname.KuasaPengguna, "|", opdname.SubKuasaPengguna, " : ", startTime.String())
	// get data
	report := []models.ResponseRekapitulasi{}
	report, _, _, _, _, err = usecase.NewReportRekapitulasiUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.REDIS.RedisCache).Export(0, 0, params.F_Periode, params.F_Penggunafilter,
		params.Penggunafilter, params.F_Kuasapengguna_Filter, params.Kuasapengguna_Filter, params.F_Subkuasa_Filter, params.Subkuasa_Filter,
		params.F_Tahun, params.F_Bulan, params.F_Jenis, params.Action, params.Firstload, params.Draw, params.F_Jenisrekap)
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
	timestr := t.DB.NowFunc().Format(REKAPITULASI_FORMAT_FILE_TIME)
	folderPath := os.Getenv("FOLDER_REPORT")
	folderReport := REKAPITULASI_EXCEL_FILE_FOLDER
	os.MkdirAll(folderPath+"/"+folderReport, os.ModePerm)
	fileName := ""
	if opdname.Pengguna == "" && opdname.KuasaPengguna == "" {
		fileName = "Rekapitulasi " + timestr
	} else {
		fileName = opdname.Pengguna + ":" + opdname.KuasaPengguna + ":" + opdname.SubKuasaPengguna + " " + timestr
	}
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
