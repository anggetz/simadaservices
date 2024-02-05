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

type reportATLUseCase struct {
	db         *gorm.DB
	redisCache *cache.Cache
}

func NewReportATLUseCase(db *gorm.DB, redisCache *cache.Cache) *reportATLUseCase {
	return &reportATLUseCase{
		db:         db,
		redisCache: redisCache,
	}
}

func (i *reportATLUseCase) Get(start int, limit int, jenis string, g *gin.Context) ([]models.ReportMDBATL, int64, int64, int64, interface{}, error) {
	// pidopd := ""
	// pidopd_cabang := ""
	// pidupt := ""

	// tglawal := ""
	// tglakhir := ""

	// if g.Query("f_periode") == "1" { // bulan
	// 	tglawal = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-01"
	// 	tglakhir = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-31"
	// } else if g.Query("f_periode") == "2" { // triwulan 1 (1 januari - 31 maret)
	// 	tglawal = g.Query("f_tahun") + "-01-01"
	// 	tglakhir = g.Query("f_tahun") + "-03-31"
	// } else if g.Query("f_periode") == "3" { // semester 1 (1 januari - 31 juni)
	// 	tglawal = g.Query("f_tahun") + "-01-01"
	// 	tglakhir = g.Query("f_tahun") + "-06-30"
	// } else if g.Query("f_periode") == "4" { // triwulan 3 (1 juli - 30 september)
	// 	tglawal = g.Query("f_tahun") + "-07-01"
	// 	tglakhir = g.Query("f_tahun") + "-09-30"
	// } else if g.Query("f_periode") == "5" { // tahun  (1 januari - 31 desember)
	// 	tglawal = g.Query("f_tahun") + "-01-01"
	// 	tglakhir = g.Query("f_tahun") + "-12-31"
	// }
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

	return i.GetData(start, limit, tgl, pidopd, pidopd_cabang, pidupt, tahun, bulan, draw, jenis)
	// return i.GetData(start, limit, tglawal, tglakhir, pidopd, pidopd_cabang, pidupt, tahun, bulan, draw, jenis)
}

// func (i *reportATLUseCase) GetData(start int, limit int, tglawal string, tglakhir string, pidopd string, pidopd_cabang string, pidupt string, tahun string, bulan string, draw string, jenis string) ([]models.ReportMDBATL, int64, int64, int64, interface{}, error) {
func (i *reportATLUseCase) GetData(start int, limit int, tgl string, pidopd string, pidopd_cabang string, pidupt string, tahun string, bulan string, draw string, jenis string) ([]models.ReportMDBATL, int64, int64, int64, interface{}, error) {
	report := []models.ReportMDBATL{}

	tahun_sk, _ := strconv.Atoi(tahun)
	tahun_sb, _ := strconv.Atoi(tahun)
	tahun_sb = tahun_sb - 1

	// pre query
	params := i.db.Raw(`select ? tahun_sekarang, ? tahun_sebelum, ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt`, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt)
	penyusutan := i.db.Raw(`select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
	sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
	from getpenyusutan(?::int, ?::int) group by 1`, tahun_sk, tahun_sb)
	pemeliharaan := i.db.Raw(`select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan where to_char(tgl, 'yyyy-mm') <= ? group by 1`, tgl)
	// organisasi := i.db.Raw("Select id, pid, nama, level from m_organisasi")

	// main query
	sqlQuery := i.db.
		Table("detil_aset_lainnya as d").
		Select(`mb.kode_akun, mb.kode_kelompok, mb.kode_jenis, mb.kode_objek, mb.kode_rincian_objek, mb.kode_sub_rincian_objek, mb.kode_sub_sub_rincian_objek,
			i.pidopd , i.pidopd_cabang, i.pidupt , mo.nama , mo.level,
			mb.nama_rek_aset nama_barang, i.kode_nibar nibar, i.noreg nomor_register, d.spesifikasi_nama_barang,
			d.spesifikasi_lainnya ,'' lokasi,i.jumlah ,
			ms.nama satuan,i.harga_satuan harga_satuan_perolehan,i.jumlah * i.harga_satuan nilai_perolehan,
			coalesce(p.biaya,0) nilai_atribusi, coalesce(i.jumlah * i.harga_satuan,0) + coalesce(p.biaya,0) nilai_perolehan_setelah_atribusi,
			CAST(coalesce(pe.penyusutan_sd_tahun_sebelumnya,0) as numeric) penyusutan_sd_tahun_sebelumnya,
			CAST(coalesce(pe.beban_penyusutan,0) as numeric) beban_penyusutan,
			CAST(coalesce(pe.penyusutan_sd_tahun_sekarang,0) as numeric) penyusutan_sd_tahun_sekarang,
			CAST(coalesce(pe.nilai_buku,0) as numeric) nilai_buku,
			'' cara_perolehan, to_char(i.tgl_perolehan, 'dd/mm/yyyy') tgl_perolehan,
			case when mo.level = 2 then
				case when kuasa.nama != '' then kuasa.nama ||' / '||subkuasa.nama||' / '||mo.nama
				when subkuasa.nama != '' then subkuasa.nama||' / '||mo.nama
				else mo.nama end
			when mo.level = 1 then
				case when subkuasa.nama != '' then subkuasa.nama ||' / '|| mo.nama
				else mo.nama end
			else
				mo.nama
			end status_penggunaan,
			coalesce(i.keterangan, d.keterangan ) keterangan , to_char(i.tgl_dibukukan, 'dd/mm/yyyy') tgl_dibukukan,
			mo.level, i.tahun_perolehan tahun, to_char(i.tgl_perolehan, 'mm') bulan`).
		Joins("CROSS JOIN (?) pr ", params).
		Joins("Join inventaris i on i.id=d.pidinventaris").
		Joins("Left Join (?) pe on pe.inventaris_id =i.id", penyusutan).
		Joins("Join m_barang mb on mb.id=i.pidbarang").
		Joins("Join m_satuan_barang ms on ms.id=i.satuan").
		Joins("Left Join (?) p on p.pidinventaris= i.id", pemeliharaan).
		Joins("Join m_organisasi mo on mo.id = i.pidopd").
		Joins("Left Join m_organisasi subkuasa on subkuasa.id=mo.pid").
		Joins("Left Join m_organisasi kuasa on kuasa.id=subkuasa.pid").
		Joins("Left Join m_organisasi pengguna on pengguna.id=kuasa.pid").
		Where(`
			(i.pidopd::text =pr.pidopd OR trim(both from pr.pidopd)='')  and
			(i.pidopd_cabang::text =pr.pidopd_cabang OR trim(both from pr.pidopd_cabang)='') and
			(i.pidupt::text =pr.pidupt OR trim(both from pr.pidupt)='') and 
			-- i.tgl_dibukukan between pr.tanggal_awal and pr.tanggal_akhir and
			TO_CHAR(i.tgl_dibukukan, 'yyyy-mm') <= pr.tanggal and
			i.deleted_at is null and
			i.draft is null`).
		Order("i.id")

	if jenis == "export" {
		// if err := sqlQuery.Offset(start).Limit(limit).Find(&report).Error; err != nil {
		if err := sqlQuery.Find(&report).Error; err != nil {
			return nil, 0, 0, 1, 0, err
		}

		return report, 0, 0, 1, 0, nil
	}

	if err := sqlQuery.Offset(start).Limit(limit).Find(&report).Error; err != nil {
		return nil, 0, 0, 1, 0, err
	}

	var countData struct {
		Total int64
	}

	// get count filtered
	strWhere := fmt.Sprintf(`
		(i.pidopd::text = '%s' OR trim(both from '%s')='')  and
		(i.pidopd_cabang::text = '%s' OR trim(both from '%s')='') and
		(i.pidupt::text = '%s' OR trim(both from '%s')='') and
		TO_CHAR(i.tgl_dibukukan, 'yyyy-mm') <= '%s' and
		i.deleted_at IS NULL AND i.draft IS NULL`, pidopd, pidopd, pidopd_cabang, pidopd_cabang, pidupt, pidupt, tgl)

	sqlCountFiltered := i.db.Table("detil_aset_lainnya d").
		Select("count(d.id) as total").
		Joins("JOIN inventaris i ON i.id = d.pidinventaris").
		Where(strWhere)

	// get from cache
	err := i.redisCache.Get(context.TODO(), "bmd-atl-count"+strWhere, &countData.Total)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, 1, 0, err
	}

	log.Println("check redis error", err, countData.Total)

	if err == cache.ErrCacheMiss || countData.Total == 0 {
		sqlTxCountFiltered := sqlCountFiltered.Scan(&countData)

		if sqlTxCountFiltered.Error != nil {
			return nil, 0, 0, 1, 0, sqlTxCountFiltered.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   "bmd-atl-count" + strWhere,
			Value: countData.Total,
			TTL:   time.Minute * 10,
		})
	}

	var countDataFiltered int64
	countDataFiltered = countData.Total

	var ndraw int64
	if draw != "" {
		ndraw, _ = strconv.ParseInt(draw, 10, 64)
	}

	summary_perpage := models.SummaryPerPage{}
	for _, np := range report {
		summary_perpage.NilaiHargaSatuan = summary_perpage.NilaiHargaSatuan + np.HargaSatuanPerolehan
		summary_perpage.NilaiPerolehan = summary_perpage.NilaiPerolehan + np.NilaiPerolehan
		summary_perpage.NilaiAtribusi = summary_perpage.NilaiAtribusi + np.NilaiAtribusi
		summary_perpage.NilaiPerolehanSetelahAtribusi = summary_perpage.NilaiPerolehanSetelahAtribusi + np.NilaiPerolehanSetelahAtribusi
		summary_perpage.NilaiPenyusutanTahun = summary_perpage.NilaiPenyusutanTahun + np.PenyusutanTahunSebelumnya
		summary_perpage.NilaiPenyusutanPeriode = summary_perpage.NilaiPenyusutanPeriode + np.PenyusutanTahunSekarang
		summary_perpage.NilaiBebanPenyusutan = summary_perpage.NilaiBebanPenyusutan + np.BebanPenyusutan
		summary_perpage.NilaiBuku = summary_perpage.NilaiBuku + np.NilaiBuku
	}

	return report, countData.Total, countDataFiltered, ndraw, summary_perpage, nil
}

func (i *reportATLUseCase) GetTotal(start int, limit int, g *gin.Context) (*models.SummaryPage, error) {
	// pidopd := ""
	// pidopd_cabang := ""
	// pidupt := ""

	// tglawal := ""
	// tglakhir := ""

	// if g.Query("f_periode") == "1" { // bulan
	// 	tglawal = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-01"
	// 	tglakhir = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-31"
	// } else if g.Query("f_periode") == "2" { // triwulan 1 (1 januari - 31 maret)
	// 	tglawal = g.Query("f_tahun") + "-01-01"
	// 	tglakhir = g.Query("f_tahun") + "-03-31"
	// } else if g.Query("f_periode") == "3" { // semester 1 (1 januari - 31 juni)
	// 	tglawal = g.Query("f_tahun") + "-01-01"
	// 	tglakhir = g.Query("f_tahun") + "-06-30"
	// } else if g.Query("f_periode") == "4" { // triwulan 3 (1 juli - 30 september)
	// 	tglawal = g.Query("f_tahun") + "-07-01"
	// 	tglakhir = g.Query("f_tahun") + "-09-30"
	// } else if g.Query("f_periode") == "5" { // tahun  (1 januari - 31 desember)
	// 	tglawal = g.Query("f_tahun") + "-01-01"
	// 	tglakhir = g.Query("f_tahun") + "-12-31"
	// }
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

	// pre query
	// params := i.db.Raw(`select ? tahun_sekarang, ? tahun_sebelum, ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt`, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt)
	penyusutan := i.db.Raw(`select inventaris_id, sum(penyusutan_sd_tahun_sekarang) penyusutan_sd_tahun_sekarang,
	sum(beban_penyusutan) beban_penyusutan, sum(nilai_buku) nilai_buku,sum(penyusutan_sd_tahun_sebelumnya) penyusutan_sd_tahun_sebelumnya
	from getpenyusutan(?::int, ?::int) group by 1`, tahun_sk, tahun_sb)
	pemeliharaan := i.db.Raw(`select pidinventaris, coalesce(sum(biaya),0) biaya from pemeliharaan where to_char(tgl, 'yyyy-mm') <= ? group by 1`, tgl)
	organisasi := i.db.Raw("Select id, pid, nama, level from m_organisasi")

	// get count filtered
	strWhere := fmt.Sprintf(`
		(i.pidopd::text = '%s' OR trim(both from '%s')='')  and
		(i.pidopd_cabang::text = '%s' OR trim(both from '%s')='') and
		(i.pidupt::text = '%s' OR trim(both from '%s')='') and
		TO_CHAR(i.tgl_dibukukan, 'yyyy-mm') <= '%s' and
		i.deleted_at IS NULL AND i.draft IS NULL`, pidopd, pidopd, pidopd_cabang, pidopd_cabang, pidupt, pidupt, tgl)

	// main query
	sqlQuery := i.db.Table("detil_aset_lainnya as d").
		Select(`count(d.id) jumlah,
			sum(i.harga_satuan) nilai_harga_satuan,
			sum(i.jumlah * i.harga_satuan) nilai_perolehan,
			sum(coalesce(p.biaya,0)) nilai_atribusi,
			sum(coalesce(i.jumlah * i.harga_satuan,0) + coalesce(p.biaya,0)) nilai_perolehan_setelah_atribusi,
			CAST(sum(coalesce(pe.penyusutan_sd_tahun_sebelumnya,0)) as numeric) nilai_penyusutan_tahun,
			CAST(sum(coalesce(pe.beban_penyusutan,0)) as numeric) nilai_beban_penyusutan,
			CAST(sum(coalesce(pe.penyusutan_sd_tahun_sekarang,0)) as numeric) nilai_penyusutan_periode,
			CAST(sum(coalesce(pe.nilai_buku,0)) as numeric) nilai_buku`).
		Joins("Join inventaris i on i.id=d.pidinventaris").
		Joins("Left Join (?) pe on pe.inventaris_id =i.id", penyusutan).
		Joins("Left Join (?) p on p.pidinventaris= i.id", pemeliharaan).
		Joins("Left Join m_organisasi mo on mo.id = i.pidopd").
		Joins("Left Join (?) subkuasa on subkuasa.id=mo.pid", organisasi).
		Joins("Left Join (?) kuasa on kuasa.id=subkuasa.pid", organisasi).
		Joins("Left Join (?) pengguna on pengguna.id=kuasa.pid", organisasi).
		Where(strWhere)

	// get from cache
	err := i.redisCache.Get(context.TODO(), "bmd-atl-total"+strWhere, &summary_page)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, err
	}

	if err == cache.ErrCacheMiss || summary_page.Jumlah == 0 {
		sqlTxFiltered := sqlQuery.Find(&summary_page)

		if sqlTxFiltered.Error != nil {
			return nil, err
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   "bmd-atl-total" + strWhere,
			Value: summary_page,
			TTL:   time.Minute * 10,
		})
	}

	return &summary_page, nil
}

func (i *reportATLUseCase) Export(start int, limit int, f_periode string, f_penggunafilter string, penggunafilter string, f_kuasapengguna_filter string, kuasapengguna_filter string, f_subkuasa_filter string, subkuasa_filter string, f_tahun string, f_bulan string, f_jenis string, action string, firstload string, draw string) ([]models.ReportMDBATL, int64, int64, int64, interface{}, error) {
	// pidopd := ""
	// pidopd_cabang := ""
	// pidupt := ""

	// tglawal := ""
	// tglakhir := ""

	// if f_periode == "1" {
	// 	tglawal = f_tahun + "-" + f_bulan + "-01"
	// 	tglakhir = f_tahun + "-" + f_bulan + "-31"
	// } else if f_periode == "2" {
	// 	tglawal = f_tahun + "-01-01"
	// 	tglakhir = f_tahun + "-03-31"
	// } else if f_periode == "3" {
	// 	tglawal = f_tahun + "-01-01"
	// 	tglakhir = f_tahun + "-06-30"
	// } else if f_periode == "4" {
	// 	tglawal = f_tahun + "-07-01"
	// 	tglakhir = f_tahun + "-09-30"
	// } else if f_periode == "5" {
	// 	tglawal = f_tahun + "-01-01"
	// 	tglakhir = f_tahun + "-12-31"
	// }
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

	return i.GetData(start, limit, tgl, pidopd, pidopd_cabang, pidupt, f_tahun, f_bulan, draw, "export")
	// return i.GetData(start, limit, tglawal, tglakhir, pidopd, pidopd_cabang, pidupt, f_tahun, f_bulan, draw, "export")
}
