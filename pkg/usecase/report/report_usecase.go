package usecase

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"path/filepath"
	"simadaservices/pkg/models"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ATL_EXCEL_FILE_FOLDER          = "bmdatl"
	TANAH_EXCEL_FILE_FOLDER        = "bmdtanah"
	REKAPITULASI_EXCEL_FILE_FOLDER = "rekapitulasi"
	MUTASIBMD_EXCEL_FILE_FOLDER    = "mutasibmd"

	FORMAT_FILE_TIME = "02-01-2006 15:04:05"
)

type reportUseCase struct {
	db *gorm.DB
}

func ReportUseCase(db *gorm.DB) *reportUseCase {
	return &reportUseCase{
		db: db,
	}
}

type OpdName struct {
	Pengguna         string `default:""`
	KuasaPengguna    string `default:""`
	SubKuasaPengguna string `default:""`
}

type QueryParams struct {
	Action                    string `form:"action"`
	F_Bulan                   string `form:"f_bulan"`
	F_Jenis                   string `form:"f_jenis"`
	F_Periode                 string `form:"f_periode"`
	F_Tahun                   string `form:"f_tahun"`
	Firstload                 string `form:"firstload"`
	F_Penggunafilter          string `form:"f_penggunafilter"`
	F_Kuasapengguna_Filter    string `form:"f_kuasapengguna_filter"`
	F_Subkuasa_Filter         string `form:"f_subkuasa_filter"`
	Penggunafilter            string `form:"penggunafilter"`
	Kuasapengguna_Filter      string `form:"kuasapengguna_filter"`
	Subkuasa_Filter           string `form:"subkuasa_filter"`
	JenisPeriode              string `form:"f_jenisperiode"`
	F_Jenisbarang_Filter      string `form:"f_jenisbarang_filter"`
	F_Kodeobjek_Filter        string `form:"f_kodeobjek_filter"`
	F_Koderincianobjek_Filter string `form:"f_koderincianobjek_filter"`
	Draw                      string `form:"draw"`
}

func (i *reportUseCase) GetOpdName(c string) OpdName {
	var params QueryParams
	opdName := OpdName{}

	err := json.Unmarshal([]byte(c), &params)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return opdName
	}

	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if params.F_Penggunafilter != "" {
		pidopd = params.F_Penggunafilter
	} else {
		if params.Penggunafilter != "" {
			pidopd = params.Penggunafilter
		}
	}

	if params.F_Kuasapengguna_Filter != "" {
		pidopd_cabang = params.F_Kuasapengguna_Filter
	} else {
		if params.Kuasapengguna_Filter != "" {
			pidopd_cabang = params.Kuasapengguna_Filter
		}
	}

	if params.F_Subkuasa_Filter != "" {
		pidupt = params.F_Subkuasa_Filter
	} else {
		if params.Subkuasa_Filter != "" {
			pidupt = params.Subkuasa_Filter
		}
	}

	opd := models.Organisasi{}
	opdcabang := models.Organisasi{}
	opdsubcabang := models.Organisasi{}

	if pidopd != "" {
		i.db.First(&opd, pidopd)
		opdName.Pengguna = opd.Nama
	}

	if pidopd_cabang != "" {
		i.db.First(&opdcabang, pidopd_cabang)
		opdName.KuasaPengguna = opdcabang.Nama
	}

	if pidupt != "" {
		i.db.First(&opdsubcabang, pidupt)
		opdName.SubKuasaPengguna = opdsubcabang.Nama
	}

	return opdName
}

func (i *reportUseCase) GetPengguna() ([]models.Organisasi, error) {
	pengguna := []models.Organisasi{}
	if err := i.db.Find(&pengguna).Where("level = ?", 0).Error; err != nil {
		return nil, err
	}
	return pengguna, nil
}

func (i *reportUseCase) GetFileExport(g *gin.Context) ([]models.FileStruct, error) {
	fileList := []models.FileStruct{}

	reportType := g.Query("type")
	log.Println(reportType)
	arr := strings.Split(reportType, "-")
	folderPath := os.Getenv("FOLDER_REPORT") + "/" + arr[2]

	task := []models.TaskQueue{}
	if err := i.db.Find(&task, "task_name = ?", reportType).Error; err != nil {
		return nil, err
	}

	for _, v := range task {
		fileInfo, err := os.Stat(v.CallbackLink + ".xlsx")
		if err == nil {
			if !strings.HasPrefix(fileInfo.Name(), ".") {
				fileList = append(fileList, models.FileStruct{
					FilePath:  folderPath,
					FileName:  fileInfo.Name(),
					FileSize:  math.Round(float64(fileInfo.Size())/(1024*1024)*100) / 100,
					CreatedAt: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
					Status:    v.Status,
				})
			}
		} else {
			filename := strings.Split(v.CallbackLink, "/")[len(strings.Split(v.CallbackLink, "/"))-1]
			if v.Status == "pending" {
				fileList = append(fileList, models.FileStruct{
					FilePath:  folderPath,
					FileName:  filename,
					FileSize:  0.0,
					CreatedAt: "",
					Status:    v.Status,
				})
			}
		}
	}

	return fileList, nil
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (i *reportUseCase) DeleteFileExport(g *gin.Context) error {
	file := g.Query("file")
	err := os.Remove(file)
	if err != nil {
		return err
	}

	return nil
}

func (i *reportUseCase) GetTotalOpd(penggunafilter string) ([]models.Organisasi, int64, error) {
	// jika opd_cabang kosong maka cek data untuk memecah export
	var totalOpd int64
	opd := []models.Organisasi{}
	// check opd have opd_cabang ?
	pid, _ := strconv.Atoi(penggunafilter)
	if err := i.db.Model(&models.Organisasi{}).Where("pid = ?", pid).
		Find(&opd).
		Count(&totalOpd).Error; err != nil {
		return nil, 0, err
	}

	return opd, totalOpd, nil
}

func (i *reportUseCase) SetRegisterQueue(g *gin.Context, reportType string) (*models.TaskQueue, error) {

	t, _ := g.Get("token_info")
	id := t.(jwt.MapClaims)["id"].(float64)
	folderPath := os.Getenv("FOLDER_REPORT")

	// set filename
	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if g.Query("f_penggunafilter") != "" {
		pidopd = g.Query("f_penggunafilter")
	} else {
		pidopd = g.Query("penggunafilter")
	}
	if g.Query("f_kuasapengguna_filter") != "" {
		pidopd_cabang = g.Query("f_kuasapengguna_filter")
	} else {
		pidopd_cabang = g.Query("kuasapengguna_filter")
	}
	if g.Query("f_subkuasa_filter") != "" {
		pidupt = g.Query("f_subkuasa_filter")
	} else {
		pidupt = g.Query("subkuasa_filter")
	}

	opdName := OpdName{}
	opd := models.Organisasi{}
	opdcabang := models.Organisasi{}
	opdsubcabang := models.Organisasi{}

	if pidopd != "" {
		i.db.First(&opd, pidopd)
		opdName.Pengguna = opd.Nama
	}

	if pidopd_cabang != "" {
		i.db.First(&opdcabang, pidopd_cabang)
		opdName.KuasaPengguna = opdcabang.Nama
	}

	if pidupt != "" {
		i.db.First(&opdsubcabang, pidupt)
		opdName.SubKuasaPengguna = opdsubcabang.Nama
	}

	username := t.(jwt.MapClaims)["username"].(string)
	fileName := opdName.Pengguna + ":" + opdName.KuasaPengguna + ":" + opdName.SubKuasaPengguna + "-" + username

	tq := models.TaskQueue{
		TaskUUID:     uuid.NewString(),
		TaskName:     "worker-export-" + reportType,
		TaskType:     "export_report",
		Status:       "pending",
		CreatedBy:    int(id),
		CallbackLink: folderPath + "/" + reportType + "/" + fileName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := i.db.Create(&tq).Error; err != nil {
		return nil, err
	}

	return &tq, nil
}
