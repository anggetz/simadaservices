package usecase

import (
	"context"
	"fmt"
	"log"
	"simadaservices/pkg/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"gorm.io/gorm"
)

type reportRekapitulasiUseCase struct {
	db         *gorm.DB
	redisCache *cache.Cache
}

func NewReportRekapitulasiUseCase(db *gorm.DB, redisCache *cache.Cache) *reportRekapitulasiUseCase {
	return &reportRekapitulasiUseCase{
		db:         db,
		redisCache: redisCache,
	}
}

func (i *reportRekapitulasiUseCase) Get(start int, limit int, g *gin.Context) ([]models.ResponseRekapitulasi, int64, int64, int64, interface{}, error) {
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

	// firstload, _ := strconv.ParseBool(g.Query("firstload"))
	if g.Query("f_penggunafilter") != "" {
		pidopd = g.Query("f_penggunafilter")
	} else {
		if g.Query("penggunafilter") != "" {
			pidopd = g.Query("penggunafilter")
		}
	}

	if g.Query("f_kuasapengguna_filter") != "" {
		pidopd_cabang = g.Query("f_kuasapengguna_filter")
	} else {
		if g.Query("kuasapengguna_filter") != "" {
			pidopd_cabang = g.Query("kuasapengguna_filter")
		}
	}

	if g.Query("f_subkuasa_filter") != "" {
		pidopd = g.Query("f_subkuasa_filter")
	} else {
		if g.Query("subkuasa_filter") != "" {
			pidupt = g.Query("subkuasa_filter")
		}
	}

	tahun := g.Query("f_tahun")
	bulan := g.Query("f_bulan")
	draw := g.Query("draw")
	jenisrekap := g.Query("f_jenisrekap")

	return i.GetData(start, limit, tgl, pidopd, pidopd_cabang, pidupt, tahun, bulan, draw, jenisrekap, "")
}

func (i *reportRekapitulasiUseCase) Export(start int, limit int, f_periode string, f_penggunafilter string, penggunafilter string, f_kuasapengguna_filter string, kuasapengguna_filter string, f_subkuasa_filter string, subkuasa_filter string, f_tahun string, f_bulan string, f_jenis string, action string, firstload string, draw string, jenisrekap string) ([]models.ResponseRekapitulasi, int64, int64, int64, interface{}, error) {
	tgl := ""
	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if f_periode == "1" {
		tgl = f_tahun + "-" + f_bulan
	} else if f_periode == "2" {
		tgl = f_tahun + "-03"
	} else if f_periode == "3" {
		tgl = f_tahun + "-06"
	} else if f_periode == "4" {
		tgl = f_tahun + "-09"
	} else if f_periode == "5" {
		tgl = f_tahun + "-" + f_bulan
	}

	if f_penggunafilter != "" {
		pidopd = f_penggunafilter
	} else {
		if penggunafilter != "" {
			pidopd = penggunafilter
		}
	}

	if f_kuasapengguna_filter != "" {
		pidopd_cabang = f_kuasapengguna_filter
	} else {
		if kuasapengguna_filter != "" {
			pidopd_cabang = kuasapengguna_filter
		}
	}

	if f_subkuasa_filter != "" {
		pidopd = f_subkuasa_filter
	} else {
		if subkuasa_filter != "" {
			pidupt = subkuasa_filter
		}
	}

	return i.GetData(start, limit, tgl, pidopd, pidopd_cabang, pidupt, f_tahun, f_bulan, draw, jenisrekap, "export")
}

func (i *reportRekapitulasiUseCase) GetData(start int, limit int, tgl string, pidopd string, pidopd_cabang string, pidupt string, tahun string, bulan string, draw string, jenisrekap string, jenis string) ([]models.ResponseRekapitulasi, int64, int64, int64, interface{}, error) {
	inventaris := []models.ResponseRekapitulasi{}

	log.Println(tgl, pidopd, pidopd_cabang, pidupt, tahun, bulan, draw, jenisrekap, jenis)

	tahun_sk, _ := strconv.Atoi(tahun)
	tahun_sb, _ := strconv.Atoi(tahun)
	tahun_sb = tahun_sb - 1

	// pre query
	params := i.db.Raw(`select ? tahun_sekarang, ? tahun_sebelum, ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt`, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt)
	penyusutan := i.db.Raw(`select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
	sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
	from getpenyusutan(?::int, ?::int) group by 1`, tahun_sk, tahun_sb)
	pemeliharaan := i.db.Raw(`select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan where to_char(tgl, 'yyyy-mm') <= ? group by 1`, tgl)

	// main query
	sqlQuery := i.db.Table("public.reportrekap as rp")

	// get the filter
	if jenisrekap == "1" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, rp.kode_jenis kode_barang, COALESCE(SUM( rp.jumlah), 0) jumlah, COALESCE(SUM( rp.nilai_perolehan), 0) nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) nilai_atribusi, COALESCE(SUM( rp.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) as NUMERIC) akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) as NUMERIC) nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis) =  rp.kode_jenis AND b.kode_objek IS NULL AND b.kode_jenis IS NOT NULL")
	} else if jenisrekap == "2" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, rp.kode_objek kode_barang, COALESCE(SUM( rp.jumlah), 0) jumlah, COALESCE(SUM( rp.nilai_perolehan), 0) nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) nilai_atribusi, COALESCE(SUM( rp.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) as NUMERIC) akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) as NUMERIC) nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek) =  rp.kode_objek AND b.kode_rincian_objek IS NULL")
	} else if jenisrekap == "3" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, rp.kode_rincian_objek kode_barang, COALESCE(SUM( rp.jumlah), 0) jumlah, COALESCE(SUM( rp.nilai_perolehan), 0) nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) nilai_atribusi, COALESCE(SUM( rp.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) as NUMERIC) akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) as NUMERIC) nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek, b.kode_rincian_objek) = rp.kode_rincian_objek AND b.kode_sub_rincian_objek IS NULL")
	} else if jenisrekap == "4" {
		sqlQuery = sqlQuery.
			Select(`b.nama_rek_aset nama_barang, rp.kode_sub_rincian_objek kode_barang, COALESCE(SUM(rp.jumlah), 0) jumlah, COALESCE(SUM(rp.nilai_perolehan), 0) nilai_perolehan, 
				COALESCE(SUM(pe.biaya), 0) nilai_atribusi, COALESCE(SUM(rp.nilai_perolehan), 0) + COALESCE(SUM(pe.biaya), 0) nilai_perolehan_setelah_atribusi, 
				CAST(SUM(COALESCE(p.penyusutan_sd_tahun_sekarang, 0)) as NUMERIC) akumulasi_penyusutan, CAST(SUM(COALESCE(p.nilai_buku, 0)) as NUMERIC) nilai_buku`).
			Joins("CROSS JOIN (?) pr", params).
			Joins("join m_barang b ON CONCAT_WS('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek, b.kode_rincian_objek, b.kode_sub_rincian_objek) = rp.kode_sub_rincian_objek AND b.kode_sub_sub_rincian_objek IS NULL")
	}

	// join and filter
	sqlQuery = sqlQuery.
		Joins("LEFT JOIN (?) p on p.inventaris_id = rp.id", penyusutan).
		Joins("left join (?) pe ON pe.pidinventaris = rp.id", pemeliharaan).
		Where(`(rp.pidopd::TEXT = pr.pidopd OR TRIM(BOTH FROM pr.pidopd) = '') 
				AND (rp.pidopd_cabang::TEXT = pr.pidopd_cabang OR TRIM(BOTH FROM pr.pidopd_cabang) = '') 
				AND (rp.pidupt::TEXT = pr.pidupt OR TRIM(BOTH FROM pr.pidupt) = '') 
				AND TO_CHAR(rp.tgl_dibukukan, 'yyyy-mm') <= pr.tanggal 
				AND rp.draft is null `)

	if jenisrekap == "1" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, rp.kode_jenis ").Order("rp.kode_jenis")
	} else if jenisrekap == "2" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, rp.kode_objek ").Order("rp.kode_objek")
	} else if jenisrekap == "3" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, rp.kode_rincian_objek ").Order("rp.kode_rincian_objek")
	} else if jenisrekap == "4" {
		sqlQuery = sqlQuery.Group("b.nama_rek_aset, rp.kode_sub_rincian_objek ").Order("rp.kode_sub_rincian_objek")
	}

	if jenis == "export" {
		// if err := sqlQuery.Offset(start).Limit(limit).Find(&report).Error; err != nil {
		if err := sqlQuery.Scan(&inventaris).Error; err != nil {
			return nil, 0, 0, 1, 0, err
		}

		return inventaris, 0, 0, 1, 0, nil
	}

	sqlQuery = sqlQuery.Offset(start).Limit(limit)

	if err := sqlQuery.Find(&inventaris).Error; err != nil {
		return nil, 0, 0, 1, 0, err
	}

	var countData int64

	// get count filtered
	strWhere := fmt.Sprintf(`
		(rp.pidopd::TEXT = '%s' OR TRIM(BOTH FROM '%s' ) = '') 
		AND (rp.pidopd_cabang::TEXT = '%s' OR TRIM(BOTH FROM '%s') = '') 
		AND (rp.pidupt::TEXT = '%s' OR TRIM(BOTH FROM '%s') = '') 
		AND TO_CHAR(rp.tgl_dibukukan, 'yyyy-mm') <= '%s' 
		AND rp.draft is null `, pidopd, pidopd, pidopd_cabang, pidopd_cabang, pidupt, pidupt, tgl)

	// get from cache
	err := i.redisCache.Get(context.TODO(), "rekapitulasi-count"+strWhere, &countData)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, 1, 0, err
	}

	if err == cache.ErrCacheMiss || countData == 0 {
		sqlTxCountFiltered := sqlQuery.Count(&countData)

		if sqlTxCountFiltered.Error != nil {
			return nil, 0, 0, 1, 0, sqlTxCountFiltered.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   "rekapitulasi-count" + strWhere,
			Value: countData,
			TTL:   time.Minute * 10,
		})
	}

	// var countDataFiltered int64
	countDataFiltered := countData

	var ndraw int64
	if draw != "" {
		ndraw, _ = strconv.ParseInt(draw, 10, 64)
	}

	summary_perpage := models.SummaryPerPage{}
	for _, np := range inventaris {
		summary_perpage.NilaiPerolehan = summary_perpage.NilaiPerolehan + np.NilaiPerolehan
		summary_perpage.NilaiAtribusi = summary_perpage.NilaiAtribusi + np.NilaiAtribusi
		summary_perpage.NilaiPerolehanSetelahAtribusi = summary_perpage.NilaiPerolehanSetelahAtribusi + np.NilaiPerolehanSetelahAtribusi
		summary_perpage.NilaiAkumulasiPenyusutan = summary_perpage.NilaiAkumulasiPenyusutan + np.AkumulasiPenyusutan
		summary_perpage.NilaiBuku = summary_perpage.NilaiBuku + np.NilaiBuku
	}

	return inventaris, countData, countDataFiltered, ndraw, summary_perpage, nil
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

	// firstload, _ := strconv.ParseBool(g.Query("firstload"))
	if g.Query("f_penggunafilter") != "" {
		pidopd = g.Query("f_penggunafilter")
	} else {
		if g.Query("penggunafilter") != "" {
			pidopd = g.Query("penggunafilter")
		}
	}

	if g.Query("f_kuasapengguna_filter") != "" {
		pidopd_cabang = g.Query("f_kuasapengguna_filter")
	} else {
		if g.Query("kuasapengguna_filter") != "" {
			pidopd_cabang = g.Query("kuasapengguna_filter")
		}
	}

	if g.Query("f_subkuasa_filter") != "" {
		pidopd = g.Query("f_subkuasa_filter")
	} else {
		if g.Query("subkuasa_filter") != "" {
			pidupt = g.Query("subkuasa_filter")
		}
	}

	tahun_sk, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb = tahun_sb - 1

	summary_page := models.SummaryPage{}

	log.Println("query >>", g.Query("f_penggunafilter"), g.Query("penggunafilter"))
	log.Println("Params", pidopd, pidopd_cabang, pidupt)

	// pre query
	// params := i.db.Raw(`select ? tahun_sekarang, ? tahun_sebelum, ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt`, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt)
	penyusutan := i.db.Raw(`select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
	sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
	from getpenyusutan(?::int, ?::int) group by 1`, tahun_sk, tahun_sb)
	pemeliharaan := i.db.Raw(`select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan where to_char(tgl, 'yyyy-mm') <= ? group by 1`, tgl)

	// get count filtered
	strWhere := fmt.Sprintf(`
		(reportrekap.pidopd::TEXT = '%s' OR TRIM(BOTH FROM '%s' ) = '') 
		AND (reportrekap.pidopd_cabang::TEXT = '%s' OR TRIM(BOTH FROM '%s') = '') 
		AND (reportrekap.pidupt::TEXT = '%s' OR TRIM(BOTH FROM '%s') = '') 
		AND TO_CHAR(reportrekap.tgl_dibukukan, 'yyyy-mm') <= '%s' 
		AND reportrekap.draft is null `, pidopd, pidopd, pidopd_cabang, pidopd_cabang, pidupt, pidupt, tgl)

	// main query
	sqlQuerySum := i.db.Table("reportrekap").
		Select(`sum(reportrekap.nilai_perolehan) nilai_perolehan,
		coalesce(sum(pe.biaya),0) nilai_atribusi, coalesce(sum(reportrekap.nilai_perolehan),0) + coalesce(sum(pe.biaya),0) nilai_perolehan_setelah_atribusi,
		CAST(sum(coalesce(p.penyusutan_sd_tahun_sekarang,0)) as numeric) nilai_akumulasi_penyusutan,
		CAST(sum(coalesce(p.nilai_buku,0)) as numeric) nilai_buku`)

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
		Where(strWhere)

	if err := sqlQuerySum.Find(&summary_page).Error; err != nil {
		return nil, err
	}

	// get from cache
	err := i.redisCache.Get(context.TODO(), "rekapitulasi-total"+strWhere, &summary_page)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, err
	}

	if err == cache.ErrCacheMiss || summary_page.Jumlah == 0 {
		sqlTxFiltered := sqlQuerySum.Find(&summary_page)

		if sqlTxFiltered.Error != nil {
			return nil, err
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   "rekapitulasi-total" + strWhere,
			Value: summary_page,
			TTL:   time.Minute * 10,
		})
	}

	return &summary_page, nil
}
