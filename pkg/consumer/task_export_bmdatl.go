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

type TaskExportBMDATL struct {
	rmq.Consumer
	Payload TaskExportBMDATLPayload
	DB      *gorm.DB
	Redis   *cache.Cache
}

type TaskExportBMDATLPayload struct {
	Headers []string
	Data    []byte
}

const (
	ATL_EXCEL_FILE_FOLDER = "bmd_atl"
	ATL_FORMAT_FILE_TIME  = "02-01-2006 15:04:05"
)

type QueryParamsAtl struct {
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
	QueueId                string `form:"queue_id"`
}

func (t *TaskExportBMDATL) Consume(d rmq.Delivery) {
	var err error

	fmt.Println("performing task report bmd atl")

	// get params
	var params QueryParamsAtl
	err = json.Unmarshal([]byte(d.Payload()), &params)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return
	}

	opdname := usecase.OpdName{}
	opdname = usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetOpdName(d.Payload())

	timestr := t.DB.NowFunc().Format(ATL_FORMAT_FILE_TIME)
	folderPath := os.Getenv("FOLDER_REPORT")
	folderReport := ATL_EXCEL_FILE_FOLDER
	fileName := opdname.Pengguna + ":" + opdname.KuasaPengguna + ":" + opdname.SubKuasaPengguna + " " + timestr

	defer func(errors error) {
		if errors != nil {
			// error
			log.Println("Error:", errors.Error())
		} else {
			// success, update sukses status task queue
			tq := models.TaskQueue{}
			t.DB.First(&tq, "id = ?", params.QueueId)
			tq.Status = "Success"
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
	// report := []models.ReportMDBATL{}
	report, _, _, _, _, _ := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.REDIS.RedisCache).Export(0, 0, params.F_Periode, params.F_Penggunafilter,
		params.Penggunafilter, params.F_Kuasapengguna_Filter, params.Kuasapengguna_Filter, params.F_Subkuasa_Filter, params.Subkuasa_Filter,
		params.F_Tahun, params.F_Bulan, params.F_Jenis, params.Action, params.Firstload, params.Draw)
	log.Println(" -->> RES DATA : ", t.DB.NowFunc().String())

	log.Println(" -->> CREATE FILE : ", t.DB.NowFunc().String())
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err.Error())
		}
	}()
	// Set the sheet name
	sheetName := "Sheet1"

	// Set header data
	headers := []string{"No", "Kode Barang", "Nama Barang", "NIBAR", "Nomor Register",
		"Spesifikasi Nama Barang", "Spesifikasi Lainnya", "Lokasi", "Jumlah", "Satuan",
		"Harga Satuan Perolehan (Rp)", "Nilai Perolehan (Rp)", "Nilai Atribusi (Rp)", "Nilai Perolehan Setelah Atribusi (Rp)",
		"Penyusutan s.d Tahun 2022 (Rp)", "Beban Penyusutan (Rp)", "Penyusutan s.d Periode Bulan 2023 (Rp)", "Nilai Buku",
		"Cara Perolehan", "Tanggal Perolehan", "Status Penggunan", "KET."}

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
		if colIndex == 1 {
			f.MergeCell(sheetName, "B1", "H1")
			f.SetCellStyle(sheetName, "B1", "H1", headerStyle)
			addData = 7
		}
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
				if colIndex == 1 {
					f.MergeCell(sheetName, "B1", "H1")
					f.SetCellStyle(sheetName, "B1", "H1", headerStyle)
					addData = 7
				}
				cellName, _ = excelize.CoordinatesToCellName(colIndex+addData, 1)
				f.SetCellValue(sheetName, cellName, header)
				f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
			}
		}

		newno = strconv.Itoa(totalRows + 1)
		f.SetCellValue(sheetName, "A"+newno, no)
		f.SetCellValue(sheetName, "B"+newno, drow.KodeAkun)
		f.SetCellValue(sheetName, "C"+newno, drow.KodeKelompok)
		f.SetCellValue(sheetName, "D"+newno, drow.KodeJenis)
		f.SetCellValue(sheetName, "E"+newno, drow.KodeObjek)
		f.SetCellValue(sheetName, "F"+newno, drow.KodeRincianObjek)
		f.SetCellValue(sheetName, "G"+newno, drow.KodeSubRincianObjek)
		f.SetCellValue(sheetName, "H"+newno, drow.KodeSubSubRincianObjek)
		f.SetCellValue(sheetName, "I"+newno, drow.NamaBarang)
		f.SetCellValue(sheetName, "J"+newno, drow.Nibar)
		f.SetCellValue(sheetName, "K"+newno, drow.NomorRegister)
		f.SetCellValue(sheetName, "L"+newno, drow.SpesifikasiNamaBarang)
		f.SetCellValue(sheetName, "M"+newno, drow.SpesifikasiLainnya)
		f.SetCellValue(sheetName, "N"+newno, drow.Lokasi)
		f.SetCellValue(sheetName, "O"+newno, drow.Jumlah)
		f.SetCellValue(sheetName, "P"+newno, drow.Satuan)
		f.SetCellValue(sheetName, "Q"+newno, drow.HargaSatuanPerolehan)
		f.SetCellValue(sheetName, "R"+newno, drow.NilaiPerolehan)
		f.SetCellValue(sheetName, "S"+newno, drow.NilaiAtribusi)
		f.SetCellValue(sheetName, "T"+newno, drow.NilaiPerolehanSetelahAtribusi)
		f.SetCellValue(sheetName, "U"+newno, drow.PenyusutanTahunSebelumnya)
		f.SetCellValue(sheetName, "V"+newno, drow.BebanPenyusutan)
		f.SetCellValue(sheetName, "W"+newno, drow.PenyusutanTahunSekarang)
		f.SetCellValue(sheetName, "X"+newno, drow.NilaiBuku)
		f.SetCellValue(sheetName, "Y"+newno, drow.CaraPerolehan)
		f.SetCellValue(sheetName, "Z"+newno, drow.TglPerolehan)
		f.SetCellValue(sheetName, "AA"+newno, drow.StatusPenggunaan)
		f.SetCellValue(sheetName, "AB"+newno, drow.Keterangan)

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
