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

type TaskExportMutasiBMD struct {
	rmq.Consumer
	Payload TaskExportBMDATLPayload
	DB      *gorm.DB
	Redis   *cache.Cache
}

type TaskExportMutasiBMDPayload struct {
	Headers []string
	Data    []byte
}

const (
	MUTASIBMD_EXCEL_FILE_FOLDER = "mutasibmd"
	MUTASIBMD_FORMAT_FILE_TIME  = "02-01-2006 15:04:05"
)

type QueryParamsMutasiBmd struct {
	Action                    string `json:"action"`
	F_Tahun                   string `json:"f_tahun"`
	F_Bulan                   string `json:"f_bulan"`
	F_Jenis                   string `json:"f_jenis"`
	F_Periode                 string `json:"f_periode"`
	Firstload                 string `json:"firstload"`
	F_Penggunafilter          string `json:"f_penggunafilter"`
	F_Kuasapengguna_Filter    string `json:"f_kuasapengguna_filter"`
	F_Subkuasa_Filter         string `json:"f_subkuasa_filter"`
	Penggunafilter            string `json:"penggunafilter"`
	Kuasapengguna_Filter      string `json:"kuasapengguna_filter"`
	Subkuasa_Filter           string `json:"subkuasa_filter"`
	Draw                      string `json:"draw"`
	F_Jenisperiode            string `json:"f_jenisperiode"`
	F_Kodejenis_Filter        string `json:"f_kode_jenis"`
	F_Kodeobjek_Filter        string `json:"f_kode_objek"`
	F_Koderincianobjek_filter string `json:"f_kode_rincian_objek"`
	QueueId                   int    `json:"queue_id"`
}

func (t *TaskExportMutasiBMD) Consume(d rmq.Delivery) {
	var err error

	fmt.Println("performing task report mutasi bmd")

	// get params
	var params QueryParamsMutasiBmd
	err = json.Unmarshal([]byte(d.Payload()), &params)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return
	}

	opdname := usecase.OpdName{}
	opdname = usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetOpdName(d.Payload())

	timestr := t.DB.NowFunc().Format(MUTASIBMD_FORMAT_FILE_TIME)
	folderPath := os.Getenv("FOLDER_REPORT")
	folderReport := MUTASIBMD_EXCEL_FILE_FOLDER
	fileName := ""
	if opdname.Pengguna != "" {
		fileName = fileName + strings.ReplaceAll(opdname.Pengguna, " ", "_")
	} else {
		fileName = "MUTASI_BMD"
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
				tq.TaskName = "worker-export-" + MUTASIBMD_EXCEL_FILE_FOLDER
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
	// report := []models.ReportMDBATL{}
	report, _, _, _, _, _ := usecase.NewReportMutasiBMDUseCase(kernel.Kernel.Config.DB.Connection, kernel.Kernel.Config.REDIS.RedisCache).Export(0, 0, params.F_Periode, params.F_Penggunafilter,
		params.Penggunafilter, params.F_Kuasapengguna_Filter, params.Kuasapengguna_Filter, params.F_Subkuasa_Filter, params.Subkuasa_Filter,
		params.F_Tahun, params.F_Bulan, params.F_Jenis, params.Action, params.Firstload, params.Draw, params.F_Jenis, params.F_Jenisperiode)
	log.Println(" -->> RES DATA : ", t.DB.NowFunc().String())

	log.Println(" -->> CREATE FILE : ", t.DB.NowFunc().String())
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err.Error())
		}
	}()

	log.Println(" -->> START INSERT DATA : ", t.DB.NowFunc().String())

	// Set the sheet name
	sheetName := "Sheet1"
	err = t.SetHeader(f, sheetName)
	if err != nil {
		log.Println("Error setting header:", err.Error())
		// Handle the error
	}

	// Set value of a cell.
	no := 1
	noSheet := 1
	totalRows := 2
	newno := ""
	for _, drow := range report {
		// fmt.Println("<<- INSERT DATA NO - ", no)

		// if sheet not enough
		if totalRows > 1048570 {
			noSheet++
			sheetName = "Sheet" + strconv.Itoa(noSheet)
			indexSheet, _ := f.NewSheet(sheetName)
			f.SetActiveSheet(indexSheet)
			totalRows = 2

			// Set the header row and make it bold
			err = t.SetHeader(f, sheetName)
			if err != nil {
				log.Println("Error setting header:", err.Error())
				// Handle the error
			}

		}

		newno = strconv.Itoa(totalRows + 1)
		f.SetCellValue(sheetName, "A"+newno, no)
		f.SetCellValue(sheetName, "B"+newno, drow.KodeBarang)
		f.SetCellValue(sheetName, "C"+newno, drow.NamaBarang)
		f.SetCellValue(sheetName, "D"+newno, drow.VolAwal)
		f.SetCellValue(sheetName, "E"+newno, drow.SaldoAwalNilaiperolehan)
		f.SetCellValue(sheetName, "F"+newno, drow.SaldoAwalAtribusi)
		f.SetCellValue(sheetName, "G"+newno, drow.SaldoAwalPerolehanatribusi)
		f.SetCellValue(sheetName, "H"+newno, drow.VolTambah)
		f.SetCellValue(sheetName, "I"+newno, drow.MutasiTambahNilaiperolehan)
		f.SetCellValue(sheetName, "J"+newno, drow.MutasiTambahAtribusi)
		f.SetCellValue(sheetName, "K"+newno, drow.MutasiTambahPerolehanatribusi)
		f.SetCellValue(sheetName, "L"+newno, drow.VolKurang)
		f.SetCellValue(sheetName, "M"+newno, drow.MutasiKurangNilaiperolehan)
		f.SetCellValue(sheetName, "N"+newno, drow.MutasiKurangAtribusi)
		f.SetCellValue(sheetName, "O"+newno, drow.MutasiKurangPerolehanatribusi)
		f.SetCellValue(sheetName, "P"+newno, drow.VolAkhir)
		f.SetCellValue(sheetName, "Q"+newno, drow.SaldoAkhirNilaiperolehan)
		f.SetCellValue(sheetName, "R"+newno, drow.SaldoAkhirAtribusi)
		f.SetCellValue(sheetName, "S"+newno, drow.SaldoAkhirPerolehanatribusi)

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

func (t *TaskExportMutasiBMD) SetHeader(f *excelize.File, sheetName string) error {
	// Set header data
	header1 := []string{"No", "Kode Barang", "Nama Barang", "Saldo Awal", "Mutasi Tambah", "Mutasi Kurang", "Saldo Akhir"}
	header2 := []string{"Vol", "Nilai Perolehan", "Atribusi", "Nilai Perolehan Setelah Atribusi"}

	// Create a bold style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Set the header row and make it bold
	f.MergeCell(sheetName, "A1", "A2")
	f.SetCellValue(sheetName, "A1", header1[0])
	f.SetCellStyle(sheetName, "A1", "A2", headerStyle)

	f.MergeCell(sheetName, "B1", "B2")
	f.SetCellValue(sheetName, "B1", header1[1])
	f.SetCellStyle(sheetName, "B1", "B2", headerStyle)

	f.MergeCell(sheetName, "C1", "C2")
	f.SetCellValue(sheetName, "C1", header1[2])
	f.SetCellStyle(sheetName, "C1", "C2", headerStyle)

	f.MergeCell(sheetName, "D1", "G1")
	f.SetCellValue(sheetName, "D1", header1[3])
	f.SetCellStyle(sheetName, "D1", "G1", headerStyle)
	f.MergeCell(sheetName, "H1", "K1")
	f.SetCellValue(sheetName, "H1", header1[4])
	f.SetCellStyle(sheetName, "H1", "K1", headerStyle)
	f.MergeCell(sheetName, "L1", "O1")
	f.SetCellValue(sheetName, "L1", header1[5])
	f.SetCellStyle(sheetName, "L1", "O1", headerStyle)
	f.MergeCell(sheetName, "P1", "S1")
	f.SetCellValue(sheetName, "P1", header1[6])
	f.SetCellStyle(sheetName, "P1", "S1", headerStyle)

	f.SetCellValue(sheetName, "D2", header2[0])
	f.SetCellStyle(sheetName, "D2", "D2", headerStyle)
	f.SetCellValue(sheetName, "E2", header2[1])
	f.SetCellStyle(sheetName, "E2", "E2", headerStyle)
	f.SetCellValue(sheetName, "F2", header2[2])
	f.SetCellStyle(sheetName, "F2", "F2", headerStyle)
	f.SetCellValue(sheetName, "G2", header2[3])
	f.SetCellStyle(sheetName, "G2", "G2", headerStyle)

	f.SetCellValue(sheetName, "H2", header2[0])
	f.SetCellStyle(sheetName, "H2", "H2", headerStyle)
	f.SetCellValue(sheetName, "I2", header2[1])
	f.SetCellStyle(sheetName, "I2", "I2", headerStyle)
	f.SetCellValue(sheetName, "J2", header2[2])
	f.SetCellStyle(sheetName, "J2", "J2", headerStyle)
	f.SetCellValue(sheetName, "K2", header2[3])
	f.SetCellStyle(sheetName, "K2", "K2", headerStyle)

	f.SetCellValue(sheetName, "L2", header2[0])
	f.SetCellStyle(sheetName, "L2", "L2", headerStyle)
	f.SetCellValue(sheetName, "M2", header2[1])
	f.SetCellStyle(sheetName, "M2", "M2", headerStyle)
	f.SetCellValue(sheetName, "N2", header2[2])
	f.SetCellStyle(sheetName, "N2", "N2", headerStyle)
	f.SetCellValue(sheetName, "O2", header2[3])
	f.SetCellStyle(sheetName, "O2", "O2", headerStyle)

	f.SetCellValue(sheetName, "P2", header2[0])
	f.SetCellStyle(sheetName, "P2", "P2", headerStyle)
	f.SetCellValue(sheetName, "Q2", header2[1])
	f.SetCellStyle(sheetName, "Q2", "Q2", headerStyle)
	f.SetCellValue(sheetName, "R2", header2[2])
	f.SetCellStyle(sheetName, "R2", "R2", headerStyle)
	f.SetCellValue(sheetName, "S2", header2[3])
	f.SetCellStyle(sheetName, "S2", "S2", headerStyle)

	return nil
}
