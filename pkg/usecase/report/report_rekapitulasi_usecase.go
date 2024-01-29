package usecase

import (
	"simadaservices/pkg/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type reportRekapitulasiUseCase struct {
	db *gorm.DB
}

func NewReportRekapitulasiUseCase(db *gorm.DB) *reportRekapitulasiUseCase {
	return &reportRekapitulasiUseCase{
		db: db,
	}
}

func (i *reportRekapitulasiUseCase) Get(start int, limit int, g *gin.Context) ([]models.ResponseRekapitulasi, int64, int64, int64, interface{}, error) {
	inventaris := []models.ResponseRekapitulasi{}

	tgl := ""
	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if g.Query("f_periode") == "1" {
		tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan")
	} else if g.Query("f_periode") == "2" {
		tgl = g.Query("f_tahun") + "-03"
	} else if g.Query("f_periode") == "3" {
		tgl = g.Query("f_tahun") + "-06"
	} else if g.Query("f_periode") == "4" {
		tgl = g.Query("f_tahun") + "-09"
	} else if g.Query("f_periode") == "5" {
		tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan")
	}

	firstload, _ := strconv.ParseBool(g.Query("firstload"))
	if firstload == true {
		if g.Query("penggunafilter") != "" {
			pidopd = g.Query("penggunafilter")
		}

		if g.Query("kuasapengguna_filter") != "" {
			pidopd_cabang = g.Query("kuasapengguna_filter")
		}

		if g.Query("subkuasa_filter") != "" {
			pidupt = g.Query("subkuasa_filter")
		}
	} else {
		if g.Query("f_penggunafilter") != "" {
			pidopd = g.Query("f_enggunafilter")
		}

		if g.Query("f_kuasapengguna_filter") != "" {
			pidopd_cabang = g.Query("f_kuasapengguna_filter")
		}

		if g.Query("f_subkuasa_filter") != "" {
			pidopd = g.Query("f_subkuasa_filter")
		}
	}

	tahun_sk, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb = tahun_sb - 1

	// pre query
	params := i.db.Raw(`select ? tahun_sekarang, ? tahun_sebelum, ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt`, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt)
	penyusutan := i.db.Raw(`select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
	sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
	from getpenyusutan(?::int, ?::int) group by 1`, tahun_sk, tahun_sb)
	pemeliharaan := i.db.Raw(`select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan where to_char(tgl, 'yyyy-mm') <= ? group by 1`, tgl)

	// main query
	sqlQuery := i.db.Table("reportrekap")

	// get the filter
	if g.Query("f_jenisrekap") == "1" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, reportrekap.kode_jenis as kode_barang, COALESCE(SUM(reportrekap.jumlah), 0) AS jumlah, COALESCE(SUM(reportrekap.nilai_perolehan), 0) AS nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) AS nilai_atribusi, COALESCE(SUM(reportrekap.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) AS nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) AS NUMERIC) AS akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) AS NUMERIC) AS nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis) = reportrekap.kode_jenis AND b.kode_objek IS NULL AND b.kode_jenis IS NOT NULL")
	} else if g.Query("jenisrekap") == "2" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, reportrekap.kode_objek as kode_barang, COALESCE(SUM(reportrekap.jumlah), 0) AS jumlah, COALESCE(SUM(reportrekap.nilai_perolehan), 0) AS nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) AS nilai_atribusi, COALESCE(SUM(reportrekap.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) AS nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) AS NUMERIC) AS akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) AS NUMERIC) AS nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek) = reportrekap.kode_objek AND b.kode_rincian_objek IS NULL")
	} else if g.Query("jenisrekap") == "3" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, reportrekap.kode_rincian_objek as kode_barang, COALESCE(SUM(reportrekap.jumlah), 0) AS jumlah, COALESCE(SUM(reportrekap.nilai_perolehan), 0) AS nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) AS nilai_atribusi, COALESCE(SUM(reportrekap.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) AS nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) AS NUMERIC) AS akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) AS NUMERIC) AS nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek, b.kode_rincian_objek) = reportrekap.kode_rincian_objek AND b.kode_sub_rincian_objek IS NULL")
	} else if g.Query("jenisrekap") == "4" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, reportrekap.kode_sub_rincian_objek as kode_barang, COALESCE(SUM(reportrekap.jumlah), 0) AS jumlah, COALESCE(SUM(reportrekap.nilai_perolehan), 0) AS nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) AS nilai_atribusi, COALESCE(SUM(reportrekap.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) AS nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) AS NUMERIC) AS akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) AS NUMERIC) AS nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek, b.kode_rincian_objek, b.kode_sub_rincian_objek) = reportrekap.kode_sub_rincian_objek AND b.kode_sub_sub_rincian_objek IS NULL")
	}

	// join and filter
	sqlQuery = sqlQuery.
		Joins("LEFT JOIN (?) p on p.inventaris_id = reportrekap.id", penyusutan).
		Joins("left join (?) pe ON pe.pidinventaris = reportrekap.id", pemeliharaan).
		Where(`(reportrekap.pidopd::TEXT = pr.pidopd OR TRIM(BOTH FROM pr.pidopd) = '') 
				AND (reportrekap.pidopd_cabang::TEXT = pr.pidopd_cabang OR TRIM(BOTH FROM pr.pidopd_cabang) = '') 
				AND (reportrekap.pidupt::TEXT = pr.pidupt OR TRIM(BOTH FROM pr.pidupt) = '') 
				AND TO_CHAR(reportrekap.tgl_dibukukan, 'yyyy-mm') <= pr.tanggal 
				AND reportrekap.draft is null `)

	if g.Query("f_jenisrekap") == "1" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, reportrekap.kode_jenis ").Order("reportrekap.kode_jenis")
	} else if g.Query("f_jenisrekap") == "2" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, reportrekap.kode_objek ").Order("reportrekap.kode_objek")
	} else if g.Query("f_jenisrekap") == "3" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, reportrekap.kode_rincian_objek ").Order("reportrekap.kode_rincian_objek")
	} else if g.Query("f_jenisrekap") == "4" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, reportrekap.kode_sub_rincian_objek ").Order("reportrekap.kode_sub_rincian_objek")
	}

	sqlQuery = sqlQuery.Offset(start).Limit(limit)

	if err := sqlQuery.Find(&inventaris).Error; err != nil {
		return nil, 0, 0, 1, 0, err
	}

	var countData int64
	sqlTxCount := sqlQuery.Count(&countData)
	if err := sqlTxCount.Error; err != nil {
		return nil, 0, 0, 1, 0, err
	}

	var countDataFiltered int64
	countDataFiltered = countData

	var draw int64
	if g.Query("draw") != "" {
		draw, _ = strconv.ParseInt(g.Query("draw"), 10, 64)
	}

	summary_perpage := models.SummaryPerPage{}
	for _, np := range inventaris {
		summary_perpage.NilaiPerolehan = summary_perpage.NilaiPerolehan + np.NilaiPerolehan
		summary_perpage.NilaiAtribusi = summary_perpage.NilaiAtribusi + np.NilaiAtribusi
		summary_perpage.NilaiPerolehanSetelahAtribusi = summary_perpage.NilaiPerolehanSetelahAtribusi + np.NilaiPerolehanSetelahAtribusi
		summary_perpage.NilaiAkumulasiPenyusutan = summary_perpage.NilaiAkumulasiPenyusutan + np.AkumulasiPenyusutan
		summary_perpage.NilaiBuku = summary_perpage.NilaiBuku + np.NilaiBuku
	}

	return inventaris, countData, countDataFiltered, draw, summary_perpage, nil
}

func (i *reportRekapitulasiUseCase) GetTotal(start int, limit int, g *gin.Context) (*models.SummaryPage, error) {
	tgl := ""
	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if g.Query("f_periode") == "1" {
		tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan")
	} else if g.Query("f_periode") == "2" {
		tgl = g.Query("f_tahun") + "-03"
	} else if g.Query("f_periode") == "3" {
		tgl = g.Query("f_tahun") + "-06"
	} else if g.Query("f_periode") == "4" {
		tgl = g.Query("f_tahun") + "-09"
	} else if g.Query("f_periode") == "5" {
		tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan")
	}

	firstload, _ := strconv.ParseBool(g.Query("firstload"))
	if firstload == true {
		if g.Query("penggunafilter") != "" {
			pidopd = g.Query("penggunafilter")
		}

		if g.Query("kuasapengguna_filter") != "" {
			pidopd_cabang = g.Query("kuasapengguna_filter")
		}

		if g.Query("subkuasa_filter") != "" {
			pidupt = g.Query("subkuasa_filter")
		}
	} else {
		if g.Query("f_penggunafilter") != "" {
			pidopd = g.Query("f_enggunafilter")
		}

		if g.Query("f_kuasapengguna_filter") != "" {
			pidopd_cabang = g.Query("f_kuasapengguna_filter")
		}

		if g.Query("f_subkuasa_filter") != "" {
			pidopd = g.Query("f_subkuasa_filter")
		}
	}

	tahun_sk, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb = tahun_sb - 1

	// pre query
	params := i.db.Raw(`select ? tahun_sekarang, ? tahun_sebelum, ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt`, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt)
	penyusutan := i.db.Raw(`select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
	sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
	from getpenyusutan(?::int, ?::int) group by 1`, tahun_sk, tahun_sb)
	pemeliharaan := i.db.Raw(`select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan where to_char(tgl, 'yyyy-mm') <= ? group by 1`, tgl)

	// main query
	sqlQuerySum := i.db.Table("reportrekap").
		Select(`sum(reportrekap.nilai_perolehan) nilai_perolehan,
		coalesce(sum(pe.biaya),0) nilai_atribusi, coalesce(sum(reportrekap.nilai_perolehan),0) + coalesce(sum(pe.biaya),0) nilai_perolehan_setelah_atribusi,
		CAST(sum(coalesce(p.penyusutan_sd_tahun_sekarang,0)) as numeric) nilai_akumulasi_penyusutan,
		CAST(sum(coalesce(p.nilai_buku,0)) as numeric) nilai_buku`).
		Joins("CROSS JOIN (?) pr ", params)

	// get the filter
	if g.Query("f_jenisrekap") == "1" {
		sqlQuerySum = sqlQuerySum.Joins("join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis)=reportrekap.kode_jenis and b.kode_objek is null ")

	} else if g.Query("jenisrekap") == "2" {
		sqlQuerySum = sqlQuerySum.Joins("join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek)=reportrekap.kode_objek and b.kode_rincian_objek is null ")

	} else if g.Query("jenisrekap") == "3" {
		sqlQuerySum = sqlQuerySum.Joins("join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek)=reportrekap.kode_rincian_objek and b.kode_sub_rincian_objek is null ")

	} else if g.Query("jenisrekap") == "4" {
		sqlQuerySum = sqlQuerySum.Joins("join m_barang as b on concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek,b.kode_sub_rincian_objek)=reportrekap.kode_sub_rincian_objek and b.kode_sub_sub_rincian_objek is null ")

	}

	// join and filter
	sqlQuerySum = sqlQuerySum.
		Joins("LEFT JOIN (?) p on p.inventaris_id = reportrekap.id", penyusutan).
		Joins("left join (?) pe ON pe.pidinventaris = reportrekap.id", pemeliharaan).
		Where(`(reportrekap.pidopd::TEXT = pr.pidopd OR TRIM(BOTH FROM pr.pidopd) = '') 
				AND (reportrekap.pidopd_cabang::TEXT = pr.pidopd_cabang OR TRIM(BOTH FROM pr.pidopd_cabang) = '') 
				AND (reportrekap.pidupt::TEXT = pr.pidupt OR TRIM(BOTH FROM pr.pidupt) = '') 
				AND TO_CHAR(reportrekap.tgl_dibukukan, 'yyyy-mm') <= pr.tanggal 
				AND reportrekap.draft is null`)

	summary_page := models.SummaryPage{}
	if err := sqlQuerySum.Find(&summary_page).Error; err != nil {
		return nil, err
	}

	return &summary_page, nil
}
