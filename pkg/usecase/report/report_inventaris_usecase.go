package usecase

import (
	"fmt"
	"simadaservices/pkg/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type reportInventarisUseCase struct {
	db *gorm.DB
}

func NewReportInventarisUseCase(db *gorm.DB) *reportInventarisUseCase {
	return &reportInventarisUseCase{
		db: db,
	}
}

func (i *reportInventarisUseCase) Get(start int, limit int, g *gin.Context) ([]models.ResponseInventaris, int64, int64, int64, interface{}, interface{}, error) {
	inventaris := []models.ResponseInventaris{}

	whereClause := []string{}

	// Query
	sql := i.db.
		Model(&models.Inventaris{}).
		Offset(start).
		Limit(limit).
		Select(`
			inventaris.*,
			coalesce(harga_satuan * jumlah,0) as nilai_perolehan,
			mb.nama_rek_aset,
			mb.kode_jenis,
			mb.kode_objek,
			mb.kode_rincian_objek,
			mo.nama as organisasi_nama,
			s.status_barang,
			coalesce(s.status_approval,'') as status_approval,
			coalesce(s.created_at,null) as tgl_waktu_sensus,
			'' as status_sensus,
			case when coalesce(s.status_approval,'') = 'STEP-1' then 'Proses Verifikasi' when coalesce(s.status_approval,'')='STEP-2' then 'Sudah disensus' else
			'Belum disensus' end status_sensus
		`).
		Joins("join m_barang as mb ON mb.id = inventaris.pidbarang").
		Joins("join m_organisasi mo ON mo.id = inventaris.pidopd").
		Joins("left join inventaris_sensus s ON s.id = inventaris.id_sensus")

	// get the filter
	if g.Query("f_draft") != "" {
		if g.Query("f_draft") == "1" {
			whereClause = append(whereClause, "inventaris.draft IS NOT NULL")
		} else {
			whereClause = append(whereClause, "inventaris.draft IS NULL")
		}
	}

	if g.Query("f_jenisbarangs_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_jenis = '%s'", g.Query("f_jenisbarangs_filter")))
	}

	if g.Query("f_kodeobjek_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_objek = '%s'", g.Query("f_kodeobjek_filter")))
	}

	if g.Query("f_koderincianobjek_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_rincian_objek = '%s'", g.Query("f_koderincianobjek_filter")))
	}

	firstload, _ := strconv.ParseBool(g.Query("firstload"))
	if firstload == true {
		if g.Query("penggunafilter") != "" {
			whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd = '%s'", g.Query("penggunafilter")))
		}

		if g.Query("kuasapengguna_filter") != "" {
			whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd_cabang = '%s'", g.Query("kuasapengguna_filter")))
		}

		if g.Query("subkuasa_filter") != "" {
			whereClause = append(whereClause, fmt.Sprintf("inventaris.pidupt = '%s'", g.Query("subkuasa_filter")))
		}
	} else {
		if g.Query("f_penggunafilter") != "" {
			whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd = '%s'", g.Query("f_penggunafilter")))
		}

		if g.Query("f_kuasapengguna_filter") != "" {
			whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd_cabang = '%s'", g.Query("f_kuasapengguna_filter")))
		}

		if g.Query("f_subkuasa_filter") != "" {
			whereClause = append(whereClause, fmt.Sprintf("inventaris.pidupt = '%s'", g.Query("f_subkuasa_filter")))
		}
	}

	if g.Query("f_status_sensus") != "" {
		st := g.Query("f_status_sensus")
		if st == "2" {
			st = "STEP-1"
		} else if st == "3" {
			st = "STEP-1"
		} else {
			st = ""
		}

		if st != "" {
			whereClause = append(whereClause, fmt.Sprintf("s.status_approval = '%s'", st))
			whereClause = append(whereClause, fmt.Sprintf("date_part('year', s.created_at) = %d", i.db.NowFunc().Year()))
		}
	}

	sql = sql.Where(strings.Join(whereClause, " AND "))

	if err := sql.Find(&inventaris).Error; err != nil {
		return nil, 0, 0, 1, 0, 0, err
	}

	var countData int64
	sqlTxCount := sql.Count(&countData)
	if err := sqlTxCount.Error; err != nil {
		return nil, 0, 0, 1, 0, 0, err
	}

	var countDataFiltered int64
	countDataFiltered = countData

	var draw int64
	if g.Query("draw") != "" {
		draw, _ = strconv.ParseInt(g.Query("draw"), 10, 64)
	}

	summary := 0.0
	for _, np := range inventaris {
		summary = summary + np.NilaiPerolehan
	}

	// QUERY TOTAL PEROLEHAN
	sqlSumPelohan := i.db.
		Raw(`SELECT sum(coalesce(harga_satuan * jumlah,0)) FROM inventaris `).
		Joins("join m_barang as mb ON mb.id = inventaris.pidbarang").
		Joins("join m_organisasi mo ON mo.id = inventaris.pidopd").
		Joins("left join inventaris_sensus s ON s.id = inventaris.id_sensus")

	var totalPerolehan float64
	sqlSumPerolehan := sqlSumPelohan.Scan(&totalPerolehan)
	if err := sqlSumPerolehan.Error; err != nil {
		return nil, 0, 0, 1, 0, 0, err
	}

	summary_perpage := models.SummaryPerPage{}
	summary_perpage.NilaiPerolehan = summary

	summary_page := models.SummaryPage{}
	summary_page.NilaiPerolehan = totalPerolehan

	return inventaris, countData, countDataFiltered, draw, summary_perpage, summary_page, nil
}
