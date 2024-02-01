package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"simadaservices/cmd/service-report/kernel"
	"simadaservices/pkg/queue"
	"simadaservices/pkg/tools"
	usecase "simadaservices/pkg/usecase/report"
	"strconv"

	"github.com/adjust/rmq/v5"
	"github.com/gin-gonic/gin"
)

type Api struct{}

func NewApi() *Api {
	return &Api{}
}

func (a *Api) GetInventaris(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	data, recordsTotal, recordsFiltered, draw, summary_perpage, summary_page, err := usecase.NewReportInventarisUseCase(kernel.Kernel.Config.DB.Connection).Get(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message:         "success get data",
		Data:            data,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Draw:            draw,
		SummaryPerPage:  summary_perpage,
		SummaryPage:     summary_page,
	})

	return
}

func (a *Api) GetRekapitulasi(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	data, recordsTotal, recordsFiltered, draw, summary_perpage, err := usecase.NewReportRekapitulasiUseCase(kernel.Kernel.Config.DB.Connection).Get(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message:         "success get data",
		Data:            data,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Draw:            draw,
		SummaryPerPage:  summary_perpage,
	})
	return
}

func (a *Api) ExportRekapitulasi(g *gin.Context) {
	connectionRedis := *kernel.Kernel.Config.REDIS.Connection
	preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_EXPORT_EXCEL_BMDATL)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	payload := map[string]interface{}{
		"f_periode":              g.Query("f_periode"),
		"f_penggunafilter":       g.Query("f_penggunafilter"),
		"penggunafilter":         g.Query("penggunafilter"),
		"f_kuasapengguna_filter": g.Query("f_kuasapengguna_filter"),
		"kuasapengguna_filter":   g.Query("kuasapengguna_filter"),
		"f_subkuasa_filter":      g.Query("f_subkuasa_filter"),
		"subkuasa_filter":        g.Query("subkuasa_filter"),
		"f_tahun":                g.Query("f_tahun"),
		"f_bulan":                g.Query("f_bulan"),
		"f_jenis":                g.Query("f_jenis"),
		"action":                 g.Query("action"),
		"firstload":              g.Query("firstload"),
		"draw":                   g.Query("draw"),
		"f_jenisrekap":           g.Query("f_jenisrekap"),
	}

	if g.Query("f_kuasapengguna_filter") != "" {
		fmt.Println(">>> have kuasa filtered")
		params, err := json.Marshal(payload)
		err = preQueueWorkerExcel.PublishBytes(params)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}
	} else {
		// pengguna, _ := usecase.ReportUseCase(kernelworker.Kernel.Config.DB.Connection).GetPengguna()
		// for _, v := range pengguna {
		// payload["f_penggunafilter"] = strconv.Itoa(v.ID)
		// payload["penggunafilter"] = strconv.Itoa(v.ID)
		// fmt.Println(">>> pengguna : ", v.ID, " => ", v.Nama)
		fmt.Println(">>> not have kuasa filtered")
		payload["penggunafilter"] = "1"
		payload["f_penggunafilter"] = "1"
		params, err := json.Marshal(payload)
		err = preQueueWorkerExcel.PublishBytes(params)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}
		// }
	}

	g.JSON(200, tools.Response{
		Message: "process exporting data ",
	})

	return
}

func (a *Api) GetTotalRekapitulasi(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	summary_page, err := usecase.NewReportRekapitulasiUseCase(kernel.Kernel.Config.DB.Connection).GetTotal(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message: "success get data",
		Data:    summary_page,
	})
	return
}

func (a *Api) GetBmdAtl(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))
	// action := g.Query("action")

	data, recordsTotal, recordsFiltered, draw, summary_perpage, err := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection).Get(start, length, "", g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message:         "success get data",
		Data:            data,
		RecordsTotal:    recordsTotal,
		RecordsFiltered: recordsFiltered,
		Draw:            draw,
		SummaryPerPage:  summary_perpage,
	})

	return
}

func (a *Api) GetBmdAtlTotalRecords(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	recordsTotal, err := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection).GetTotalRecords(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message: "success get data",
		Data:    recordsTotal,
	})
	return
}

func (a *Api) GetTotalBmdAtl(g *gin.Context) {
	start, _ := strconv.Atoi(g.Query("start"))
	length, _ := strconv.Atoi(g.Query("length"))

	summary_page, err := usecase.NewReportATLUseCase(kernel.Kernel.Config.DB.Connection).GetTotal(start, length, g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	g.JSON(200, tools.HttpResponseReport{
		Message: "success get data",
		Data:    summary_page,
	})
	return
}

func (a *Api) ExportBmdAtl(g *gin.Context) {
	connectionRedis := *kernel.Kernel.Config.REDIS.Connection
	preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_EXPORT_EXCEL_BMDATL)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}

	payload := map[string]interface{}{
		"f_periode":              g.Query("f_periode"),
		"f_penggunafilter":       g.Query("f_penggunafilter"),
		"penggunafilter":         g.Query("penggunafilter"),
		"f_kuasapengguna_filter": g.Query("f_kuasapengguna_filter"),
		"kuasapengguna_filter":   g.Query("kuasapengguna_filter"),
		"f_subkuasa_filter":      g.Query("f_subkuasa_filter"),
		"subkuasa_filter":        g.Query("subkuasa_filter"),
		"f_tahun":                g.Query("f_tahun"),
		"f_bulan":                g.Query("f_bulan"),
		"f_jenis":                g.Query("f_jenis"),
		"action":                 g.Query("action"),
		"firstload":              g.Query("firstload"),
		"draw":                   g.Query("draw"),
	}

	if g.Query("f_kuasapengguna_filter") != "" {
		fmt.Println(">>> have kuasa filtered")
		params, err := json.Marshal(payload)
		err = preQueueWorkerExcel.PublishBytes(params)
		if err != nil {
			g.JSON(400, err.Error())
			g.Abort()
			return
		}
	} else {
		pengguna, _ := usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetPengguna()
		for _, v := range pengguna {
			g.Set("f_penggunafilter", strconv.Itoa(v.ID))
			opd, total, err := usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetTotalOpd(g)

			if err != nil {
				g.JSON(400, err.Error())
				g.Abort()
				return
			}
			// check opd have opd_cabang ?
			if total > 0 {
				fmt.Println(">>> have kuasa loop data", total)

				for _, v := range opd {
					payload["f_kuasapengguna_filter"] = strconv.Itoa(v.ID)
					payload["kuasapengguna_filter"] = strconv.Itoa(v.ID)
					fmt.Println(">>> kuasa : ", v.ID, " => ", v.Nama)

					params, err := json.Marshal(payload)
					err = preQueueWorkerExcel.PublishBytes(params)
					if err != nil {
						g.JSON(400, err.Error())
						g.Abort()
						return
					}
				}
			} else {
				fmt.Println(">>> just data opd")

				params, err := json.Marshal(payload)
				err = preQueueWorkerExcel.PublishBytes(params)
				if err != nil {
					g.JSON(400, err.Error())
					g.Abort()
					return
				}
			}
		}
	}

	g.JSON(200, tools.Response{
		Message: "process exporting data ",
	})

	return
}

func (a *Api) FileListExport(g *gin.Context) {
	files, err := usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetFileExport(g)
	if err != nil {
		g.JSON(404, tools.NotOkResponse{
			Message: "failed get file list",
		})

		return
	}

	g.JSON(200, tools.Response{
		Data:    files,
		Message: "process exporting data ",
	})

	return
}

func (a *Api) GetFileExport(g *gin.Context) {
	// filePath := g.Query("file")
	filePath := "storage/reports/BMD_ATL/BMD_ATL_2024-01-26 15:35:55 WIB.xlsx"

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}
	defer file.Close()
	fileInfo, _ := os.Stat(filePath)

	// Set response headers
	g.Header("Content-Disposition", "attachment; filename="+file.Name())
	g.Header("Content-Type", "application/octet-stream")
	g.Header("Content-Length", fmt.Sprint(fileInfo.Size()))

	// Copy the file content to the response writer
	_, err = io.Copy(g.Writer, file)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}
}

func (a *Api) DeleteFileExport(g *gin.Context) {
	err := usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).DeleteFileExport(g)
	if err != nil {
		g.JSON(400, err.Error())
		g.Abort()
		return
	}
	g.JSON(200, tools.Response{
		Message: "success deleted data ",
	})

	return
}

func (a *Api) CronExportBmdAtl(connection rmq.Connection, params map[string]interface{}) {
	connectionRedis := connection
	preQueueWorkerExcel, err := connectionRedis.OpenQueue(queue.QUEUE_EXPORT_EXCEL_BMDATL)
	if err != nil {
		log.Println("Error open queue : ", err.Error())
		return
	}

	payload := map[string]interface{}{
		"f_periode":              params["f_periode"],
		"f_penggunafilter":       params["f_penggunafilter"],
		"penggunafilter":         params["penggunafilter"],
		"f_kuasapengguna_filter": params["f_kuasapengguna_filter"],
		"kuasapengguna_filter":   params["kuasapengguna_filter"],
		"f_subkuasa_filter":      params["f_subkuasa_filter"],
		"subkuasa_filter":        params["subkuasa_filter"],
		"f_tahun":                params["f_tahun"],
		"f_bulan":                params["f_bulan"],
		"f_jenis":                params["f_jenis"],
		"action":                 params["action"],
		"firstload":              params["firstload"],
		"draw":                   params["draw"],
	}

	if params["f_kuasapengguna_filter"] != "" {
		fmt.Println(">>> have kuasa filtered")
		params, err := json.Marshal(payload)
		err = preQueueWorkerExcel.PublishBytes(params)
		if err != nil {
			log.Println("Error publish : ", err.Error())
			return
		}
	} else {
		pengguna, _ := usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetPengguna()
		for _, v := range pengguna {
			var g gin.Context
			g.Set("f_penggunafilter", strconv.Itoa(v.ID))
			opd, total, err := usecase.ReportUseCase(kernel.Kernel.Config.DB.Connection).GetTotalOpd(&g)

			if err != nil {
				log.Println("Error get opd : ", err.Error())
				return
			}
			// check opd have opd_cabang ?
			if total > 0 {
				fmt.Println(">>> have kuasa loop data", total)

				for _, v := range opd {
					payload["f_kuasapengguna_filter"] = strconv.Itoa(v.ID)
					payload["kuasapengguna_filter"] = strconv.Itoa(v.ID)
					fmt.Println(">>> kuasa : ", v.ID, " => ", v.Nama)

					params, err := json.Marshal(payload)
					err = preQueueWorkerExcel.PublishBytes(params)
					if err != nil {
						log.Println("Error publish : ", err.Error())
						return
					}
				}
			} else {
				fmt.Println(">>> just data opd")

				params, err := json.Marshal(payload)
				err = preQueueWorkerExcel.PublishBytes(params)
				if err != nil {
					log.Println("Error publish : ", err.Error())
					return
				}
			}
		}
	}

	return
}
