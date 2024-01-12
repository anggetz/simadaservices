package usecase

import (
	"fmt"
	"simadaservices/pkg/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type reportUseCase struct {
	db *gorm.DB
}

func NewReportUseCase(db *gorm.DB) *reportUseCase {
	return &reportUseCase{
		db: db,
	}
}

type ResponseInventaris struct {
	*models.Inventaris
	NilaiPerolehan   float64 `json:"nilai_perolehan"`
	NamaRekAset      string  `json:"nama_barang"`
	KodeJenis        string  `json:"kode_jenis"`
	KodeObjek        string  `json:"kode_objek"`
	KodeRincianObjek string  `json:"kode_rincian_objek"`
	OrganisasiNama   string  `json:"pengguna_barang"`
	StatusBarang     string  `json:"status_barang"`
	StatusApproval   string  `json:"status_approval"`
	TglWaktuSensus   string  `json:"tglwaktusensus"`
	StatusSensus     string  `json:"status_sensus"`
}

type SummaryPerPage struct {
	NilaiPerolehan                float64 `json:"nilai_perolehan"`
	NilaiAtribusi                 float64 `json:"nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64 `json:"total_nilai_perolehan_setelah_atribusi"`
	NilaiAkumulasiPenyusutan      float64 `json:"nilai_akumulasi_penyusutan"`
	NilaiBuku                     float64 `json:"nilai_buku"`
}

type SummaryPage struct {
	NilaiPerolehan                float64 `json:"total_nilai_perolehan"`
	NilaiAtribusi                 float64 `json:"total_nilai_atribusi"`
	NilaiPerolehanSetelahAtribusi float64 `json:"total_nilai_perolehan_setelah_atribusi"`
	NilaiAkumulasiPenyusutan      float64 `json:"total_nilai_akumulasi_penyusutan"`
	NilaiBuku                     float64 `json:"total_nilai_buku"`
}

type ResponseRekapitulasi struct {
	NamaBarang                    string
	KodeBarang                    string
	Jumlah                        int64
	NilaiPerolehan                float64
	NilaiAtribusi                 float64
	NilaiPerolehanSetelahAtribusi float64
	AkumulasiPenyusutan           float64
	NilaiBuku                     float64
}

func (i *reportUseCase) GetInventaris(start int, limit int, g *gin.Context) ([]ResponseInventaris, int64, int64, int64, interface{}, interface{}, error) {
	inventaris := []ResponseInventaris{}

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

	summary_perpage := SummaryPerPage{}
	summary_perpage.NilaiPerolehan = summary

	summary_page := SummaryPage{}
	summary_page.NilaiPerolehan = totalPerolehan

	return inventaris, countData, countDataFiltered, draw, summary_perpage, summary_page, nil
}

func (i *reportUseCase) GetRekapitulasi(start int, limit int, g *gin.Context) ([]ResponseRekapitulasi, int64, int64, int64, interface{}, interface{}, error) {
	inventaris := []ResponseRekapitulasi{}

	// whereClause := []string{}
	fmt.Println(g.Query("filters"))

	// Query
	sqlText := `
		WITH params as (
			select ? tahun_sekarang, ? tahun_sebelum,
			?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt
		), penyusutan as (
			select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
			sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
			from getpenyusutan((select tahun_sekarang from params)::int, (select tahun_sebelum from params)::int) group by 1
		), pemeliharaan as (
			select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan
			cross join params p
			where to_char(tgl, 'yyyy-mm') <= p.tanggal
			group by 1
		) --select * from params 
		`

	// get the filter
	if g.Query("jenisrekap") == "1" {
		sqlText = sqlText + `
					select b.nama_rek_aset nama_barang,d.kode_jenis as kode_barang,coalesce(sum(d.jumlah),0) jumlah, coalesce(sum(d.nilai_perolehan),0) nilai_perolehan,
					coalesce(sum(pe.biaya),0) nilai_atribusi, coalesce(sum(d.nilai_perolehan),0) + coalesce(sum(pe.biaya),0) nilai_perolehan_setelah_atribusi,
					CAST(sum(coalesce(p.penyusutan_sd_tahun_sekarang,0)) as numeric) akumulasi_penyusutan,CAST(sum(coalesce(p.nilai_buku,0)) as numeric) nilai_buku
					from reportrekap d
					cross join params as pr
					inner join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis)=d.kode_jenis and b.kode_objek is null and b.kode_jenis is not null
				`
	} else if g.Query("jenisrekap") == "2" {
		sqlText = sqlText + `
					select b.nama_rek_aset nama_barang,d.kode_jenis as kode_barang,coalesce(sum(d.jumlah),0) jumlah, coalesce(sum(d.nilai_perolehan),0) nilai_perolehan,
					coalesce(sum(pe.biaya),0) nilai_atribusi, coalesce(sum(d.nilai_perolehan),0) + coalesce(sum(pe.biaya),0) nilai_perolehan_setelah_atribusi,
					CAST(sum(coalesce(p.penyusutan_sd_tahun_sekarang,0)) as numeric) akumulasi_penyusutan,CAST(sum(coalesce(p.nilai_buku,0)) as numeric) nilai_buku
					from reportrekap d
					cross join params as pr
					inner join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis)=d.kode_jenis and b.kode_objek is null and b.kode_jenis is not null
				`
	} else if g.Query("jenisrekap") == "3" {
		sqlText = sqlText + `
					select b.nama_rek_aset nama_barang,d.kode_rincian_objek as kode_barang,coalesce(sum(d.jumlah),0) jumlah, coalesce(sum(d.nilai_perolehan),0) nilai_perolehan,
					coalesce(sum(pe.biaya),0) nilai_atribusi, coalesce(sum(d.nilai_perolehan),0) + coalesce(sum(pe.biaya),0) nilai_perolehan_setelah_atribusi,
					CAST(sum(coalesce(p.penyusutan_sd_tahun_sekarang,0)) as numeric) akumulasi_penyusutan,CAST(sum(coalesce(p.nilai_buku,0)) as numeric) nilai_buku
					from reportrekap d
					cross join params as pr
					inner join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek)=d.kode_rincian_objek and b.kode_sub_rincian_objek is null
				`
	} else if g.Query("jenisrekap") == "4" {
		sqlText = sqlText + `
					select b.nama_rek_aset nama_barang,d.kode_sub_rincian_objek as kode_barang,coalesce(sum(d.jumlah),0) jumlah, coalesce(sum(d.nilai_perolehan),0) nilai_perolehan,
					coalesce(sum(pe.biaya),0) nilai_atribusi, coalesce(sum(d.nilai_perolehan),0) + coalesce(sum(pe.biaya),0) nilai_perolehan_setelah_atribusi,
					CAST(sum(coalesce(p.penyusutan_sd_tahun_sekarang,0)) as numeric) akumulasi_penyusutan,CAST(sum(coalesce(p.nilai_buku,0)) as numeric) nilai_buku
					from reportrekap d
					cross join params as pr
					inner join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek,b.kode_sub_rincian_objek)=d.kode_sub_rincian_objek and b.kode_sub_sub_rincian_objek is null
				`
	}

	sqlText = sqlText + ` left join penyusutan as p on p.inventaris_id=d.id
				left join pemeliharaan pe on pe.pidinventaris= d.id
				where
				(d.pidopd::text =pr.pidopd OR trim(both from pr.pidopd)='')  and
				(d.pidopd_cabang::text =pr.pidopd_cabang OR trim(both from pr.pidopd_cabang)='') and
				(d.pidupt::text =pr.pidupt OR trim(both from pr.pidupt)='')
				and to_char(d.tgl_dibukukan, 'yyyy-mm') <= pr.tanggal
				group by b.nama_rek_aset`
	tgl := ""
	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if g.Query("periode") == "1" {
		tgl = g.Query("tahun") + "-" + g.Query("bulan")
	} else if g.Query("periode") == "2" {
		tgl = g.Query("tahun") + "-03"
	} else if g.Query("periode") == "3" {
		tgl = g.Query("tahun") + "-06"
	} else if g.Query("periode") == "4" {
		tgl = g.Query("tahun") + "-09"
	} else if g.Query("periode") == "5" {
		tgl = g.Query("tahun") + "-" + g.Query("bulan")
	}

	if g.Query("pidopd") != "" {
		pidopd = g.Query("pidopd")
	}

	if g.Query("pidopd_cabang") != "" {
		pidopd_cabang = g.Query("pidopd_cabang")
	}

	if g.Query("pidupt") != "" {
		pidupt = g.Query("pidupt")
	}

	tahun_sk, _ := strconv.Atoi(g.Query("tahun"))
	tahun_sb, _ := strconv.Atoi(g.Query("tahun"))
	tahun_sb = tahun_sb - 1

	if g.Query("jenisrekap") == "1" {
		sqlText = sqlText + `, d.kode_jenis order by d.kode_jenis`
	} else if g.Query("jenisrekap") == "2" {
		sqlText = sqlText + `, d.kode_objek order by d.kode_objek`
	} else if g.Query("jenisrekap") == "3" {
		sqlText = sqlText + `, d.kode_rincian_objek order by d.kode_rincian_objek`
	} else if g.Query("jenisrekap") == "4" {
		sqlText = sqlText + `, d.kode_sub_rincian_objek order by d.kode_sub_rincian_objek`
	}

	if err := i.db.Raw(sqlText, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt).Offset(start).Limit(limit).Scan(&inventaris).Error; err != nil {
		return nil, 0, 0, 1, 0, 0, err
	}

	var countData int64
	sqlTxCount := i.db.Raw(sqlText, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt).Offset(start).Limit(limit).Count(&countData)
	if err := sqlTxCount.Error; err != nil {
		return nil, 0, 0, 1, 0, 0, err
	}

	var countDataFiltered int64
	// countDataFiltered = countData

	var draw int64
	if g.Query("draw") != "" {
		draw, _ = strconv.ParseInt(g.Query("draw"), 10, 64)
	}

	summary := 0.0
	summary_perpage := SummaryPerPage{}
	for _, np := range inventaris {
		summary_perpage.NilaiPerolehan = +np.NilaiPerolehan
		summary_perpage.NilaiAtribusi = +np.NilaiAtribusi
		summary_perpage.NilaiPerolehanSetelahAtribusi = +np.NilaiPerolehanSetelahAtribusi
		summary_perpage.NilaiAkumulasiPenyusutan = +np.AkumulasiPenyusutan
		summary_perpage.NilaiBuku = +np.NilaiBuku
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

	summary_page := SummaryPage{}
	summary_page.NilaiPerolehan = summary

	return inventaris, countData, countDataFiltered, draw, summary_perpage, summary_page, nil
}
