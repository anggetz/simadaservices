package usecase

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"path/filepath"
	"simadaservices/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	if err := i.db.Find(&pengguna).Where("level = ?", "0").Error; err != nil {
		return nil, err
	}
	return pengguna, nil
}

func (i *reportUseCase) GetFileExport(g *gin.Context) ([]models.FileStruct, error) {
	fileList := []models.FileStruct{}

	reportType := g.Query("type")
	folderPath := os.Getenv("FOLDER_REPORT") + "/" + reportType
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, err
	}
	files, err := FilePathWalkDir(folderPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileInfo, _ := os.Stat(file)
		fileList = append(fileList, models.FileStruct{
			FilePath:  folderPath,
			FileName:  fileInfo.Name(),
			FileSize:  math.Round(float64(fileInfo.Size())/(1024*1024)*100) / 100,
			CreatedAt: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
		})
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

func (i *reportUseCase) GetTotalOpd(g *gin.Context) ([]models.Organisasi, int64, error) {
	// jika opd_cabang kosong maka cek data untuk memecah export
	var totalOpd int64
	opd := []models.Organisasi{}
	// check opd have opd_cabang ?
	if err := i.db.Model(&models.Organisasi{}).Where("pid = ?", g.Query("f_penggunafilter")).
		Find(&opd).
		Count(&totalOpd).Error; err != nil {
		return nil, 0, err
	}

	return opd, totalOpd, nil
}
