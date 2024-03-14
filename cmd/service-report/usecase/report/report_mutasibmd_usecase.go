package usecase

import (
	"context"
	"fmt"
	"libcore/models"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"gorm.io/gorm"
)

type reportMutasiBMDUseCase struct {
	db         *gorm.DB
	redisCache *cache.Cache
}

func NewReportMutasiBMDUseCase(db *gorm.DB, redisCache *cache.Cache) *reportMutasiBMDUseCase {
	return &reportMutasiBMDUseCase{
		db:         db,
		redisCache: redisCache,
	}
}

type Params struct {
	Start            int
	Limit            int
	Tgl              string
	Pidopd           string
	PidopdCabang     string
	Pidupt           string
	Tglakhir         string
	Jenisperiode     string
	Tahun            string
	Bulan            string
	Draw             string
	Jenisrekap       string
	KodeJenis        string
	KodeObjek        string
	KodeRincianObjek string
	Export           string
}

func (i *reportMutasiBMDUseCase) Get(start int, limit int, g *gin.Context) ([]models.ReportMutasiBMD, int64, int64, int64, interface{}, error) {
	params := Params{}

	if g.Query("f_jenisperiode") == "1" { // semua
		if g.Query("f_periode") == "1" {
			params.Tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan")
		} else if g.Query("f_periode") == "2" {
			params.Tgl = g.Query("f_tahun") + "-03"
		} else if g.Query("f_periode") == "3" {
			params.Tgl = g.Query("f_tahun") + "-06"
		} else if g.Query("f_periode") == "4" {
			params.Tgl = g.Query("f_tahun") + "-09"
		} else if g.Query("f_periode") == "5" {
			params.Tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan")
		}
	} else { // range
		if g.Query("f_periode") == "1" { // bulan
			params.Tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-01"
			params.Tglakhir = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-31"
		} else if g.Query("f_periode") == "2" { // triwulan 1 (1 januari - 31 maret)
			params.Tgl = g.Query("f_tahun") + "-01-01"
			params.Tglakhir = g.Query("f_tahun") + "-03-31"
		} else if g.Query("f_periode") == "3" { // semester 1 (1 januari - 31 juni)
			params.Tgl = g.Query("f_tahun") + "-01-01"
			params.Tglakhir = g.Query("f_tahun") + "-06-30"
		} else if g.Query("f_periode") == "4" { // triwulan 3 (1 juli - 30 september)
			params.Tgl = g.Query("f_tahun") + "-07-01"
			params.Tglakhir = g.Query("f_tahun") + "-09-30"
		} else if g.Query("f_periode") == "5" { // tahun  (1 januari - 31 desember)
			params.Tgl = g.Query("f_tahun") + "-01-01"
			params.Tglakhir = g.Query("f_tahun") + "-12-31"
		}
	}

	// firstload, _ := strconv.ParseBool(g.Query("firstload"))
	if g.Query("f_penggunafilter") != "" {
		params.Pidopd = g.Query("f_penggunafilter")
	} else {
		if g.Query("penggunafilter") != "" {
			params.Pidopd = g.Query("penggunafilter")
		}
	}

	if g.Query("f_kuasapengguna_filter") != "" {
		params.PidopdCabang = g.Query("f_kuasapengguna_filter")
	} else {
		if g.Query("kuasapengguna_filter") != "" {
			params.PidopdCabang = g.Query("kuasapengguna_filter")
		}
	}

	if g.Query("f_subkuasa_filter") != "" {
		params.Pidupt = g.Query("f_subkuasa_filter")
	} else {
		if g.Query("subkuasa_filter") != "" {
			params.Pidupt = g.Query("subkuasa_filter")
		}
	}

	params.Tahun = g.Query("f_tahun")
	params.Bulan = g.Query("f_bulan")
	params.Draw = g.Query("draw")
	params.Jenisrekap = g.Query("f_jenisrekap")
	params.Jenisperiode = g.Query("f_jenisperiode")
	params.KodeJenis = g.Query("f_jenisbarangs_filter")
	params.KodeObjek = g.Query("f_kodeobjek_filter")
	params.KodeRincianObjek = g.Query("f_koderincianobjek_filter")
	params.Export = ""
	params.Start = start
	params.Limit = limit

	return i.GetData(params)
}

func (i *reportMutasiBMDUseCase) Export(start int, limit int, f_periode string, f_penggunafilter string, penggunafilter string, f_kuasapengguna_filter string, kuasapengguna_filter string, f_subkuasa_filter string, subkuasa_filter string, f_tahun string, f_bulan string, f_jenis string, action string, firstload string, draw string, jenisrekap string, jenisperiode string) ([]models.ReportMutasiBMD, int64, int64, int64, interface{}, error) {
	params := Params{}

	if jenisperiode == "1" {
		if f_periode == "1" {
			params.Tgl = f_tahun + "-" + f_bulan
		} else if f_periode == "2" {
			params.Tgl = f_tahun + "-03"
		} else if f_periode == "3" {
			params.Tgl = f_tahun + "-06"
		} else if f_periode == "4" {
			params.Tgl = f_tahun + "-09"
		} else if f_periode == "5" {
			params.Tgl = f_tahun + "-" + f_bulan
		}
	} else {
		if f_periode == "1" {
			params.Tgl = f_tahun + "-" + f_bulan + "-01"
			params.Tglakhir = f_tahun + "-" + f_bulan + "-31"
		} else if f_periode == "2" {
			params.Tgl = f_tahun + "-01-01"
			params.Tglakhir = f_tahun + "-03-31"
		} else if f_periode == "3" {
			params.Tgl = f_tahun + "-01-01"
			params.Tglakhir = f_tahun + "-06-30"
		} else if f_periode == "4" {
			params.Tgl = f_tahun + "-07-01"
			params.Tglakhir = f_tahun + "-09-30"
		} else if f_periode == "5" {
			params.Tgl = f_tahun + "-01-01"
			params.Tglakhir = f_tahun + "-12-31"
		}
	}

	if f_penggunafilter != "" {
		params.Pidopd = f_penggunafilter
	} else {
		if penggunafilter != "" {
			params.Pidopd = penggunafilter
		}
	}

	if f_kuasapengguna_filter != "" {
		params.PidopdCabang = f_kuasapengguna_filter
	} else {
		if kuasapengguna_filter != "" {
			params.Pidopd = kuasapengguna_filter
		}
	}

	if f_subkuasa_filter != "" {
		params.Pidupt = f_subkuasa_filter
	} else {
		if subkuasa_filter != "" {
			params.Pidupt = subkuasa_filter
		}
	}

	params.Jenisperiode = jenisperiode
	params.Jenisrekap = jenisrekap
	params.Tahun = f_tahun
	params.Bulan = f_bulan
	params.Draw = draw
	params.KodeObjek = ""
	params.KodeJenis = ""
	params.KodeRincianObjek = ""
	params.Export = "export"
	params.Start = start
	params.Limit = limit

	return i.GetData(params)
}

func (i *reportMutasiBMDUseCase) GetData(params Params) ([]models.ReportMutasiBMD, int64, int64, int64, interface{}, error) {
	inventaris := []models.ReportMutasiBMD{}

	tahun_sk, _ := strconv.Atoi(params.Tahun)
	tahun_sb, _ := strconv.Atoi(params.Tahun)
	tahun_sb = tahun_sb - 1
	month := fmt.Sprintf("%02d", time.Now().Month())

	// create view pemeliharaan
	if i.db.Migrator().HasTable("view_pemeliharaan_" + month) {
		// If the table exists, drop it
		i.db.Migrator().DropTable("view_pemeliharaan_" + month)
	}

	sqlView := fmt.Sprintf(`with params as (
				select $1::int tahun_sekarang, $2::int tahun_sebelum, $3::text tanggal,
				$4::text pidopd, $5::text pidopd_cabang, $6::text pidupt, $7::text jenis,
				$8::text kode_jenis, $9::text kode_objek , $10::text kode_rincian_objek
			), pelihara as (
				select
				pidinventaris ,
				sum(case when to_char(pm.tgl,'yyyy-mm') < p.tanggal then pm.biaya else 0 end) saldo_awal_atribusi,
				sum(case when to_char(pm.tgl,'yyyy-mm') >= p.tanggal then pm.biaya else 0 end) mutasi_tambah_atribusi
				from pemeliharaan pm
				cross join params p
				where to_char(pm.tgl, 'yyyy-mm') <= p.tanggal
				group by 1
			) select * into view_pemeliharaan_%s from pelihara;`, month)

	// Execute the query with the appropriate parameters
	i.db.Exec(sqlView, tahun_sk, tahun_sb, params.Tgl, params.Pidopd, params.PidopdCabang, params.Pidupt, params.Jenisrekap, params.KodeJenis, params.KodeObjek, params.KodeRincianObjek)

	query := `
		WITH params AS (
			select ?::int tahun_sekarang, ?::int tahun_sebelum, ?::text tanggal,
			?::text pidopd, ?::text pidopd_cabang, ?::text pidupt, ?::text jenis,
			?::text kode_jenis, ?::text kode_objek , ?::text kode_rincian_objek
		), list AS (
			SELECT
				b.nama_rek_aset AS nama_barang,
				CASE
					WHEN p.jenis = '1' THEN r.kode_jenis
					WHEN p.jenis = '2' THEN r.kode_objek
					WHEN p.jenis = '3' THEN r.kode_rincian_objek
					WHEN p.jenis = '4' THEN r.kode_sub_rincian_objek
				END AS kode_barang,
				CASE
					WHEN to_char(r.tgl_dibukukan, 'yyyy')::int < p.tahun_sekarang THEN r.jumlah
					ELSE 0
				END AS vol_awal,
				CASE
					WHEN to_char(r.tgl_dibukukan, 'yyyy')::int < p.tahun_sekarang THEN r.jumlah * r.harga_satuan
					ELSE 0
				END AS saldo_awal_nilaiperolehan,
				COALESCE(pm.saldo_awal_atribusi, 0) AS saldo_awal_atribusi,
				CASE
					WHEN to_char(r.tgl_dibukukan, 'yyyy')::int >= p.tahun_sekarang THEN r.jumlah
					ELSE 0
				END AS vol_tambah,
				CASE
					WHEN to_char(r.tgl_dibukukan, 'yyyy')::int >= p.tahun_sekarang THEN r.jumlah * r.harga_satuan
					ELSE 0
				END AS mutasi_tambah_nilaiperolehan,
				COALESCE(pm.mutasi_tambah_atribusi, 0) AS mutasi_tambah_atribusi,
				0 AS vol_kurang,
				0 AS mutasi_kurang_nilaiperolehan,
				0 AS mutasi_kurang_atribusi,
				0 AS vol_akhir,
				0 AS saldo_akhir_nilaiperolehan,
				0 AS saldo_akhir_atribusi
			FROM reportrekap r
			CROSS JOIN params p
			LEFT JOIN view_pemeliharaan_` + month + ` pm ON pm.pidinventaris = r.id
			INNER JOIN m_barang AS b ON
				CASE
					WHEN p.jenis = '1' THEN concat_ws('.', b.kode_akun, b.kode_kelompok, b.kode_jenis) = r.kode_jenis AND b.kode_objek IS NULL AND b.kode_jenis IS NOT NULL
					WHEN p.jenis = '2' THEN concat_ws('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek) = r.kode_objek AND b.kode_rincian_objek IS NULL
					WHEN p.jenis = '3' THEN concat_ws('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek, b.kode_rincian_objek) = r.kode_rincian_objek AND b.kode_sub_rincian_objek IS NULL
					WHEN p.jenis = '4' THEN concat_ws('.', b.kode_akun, b.kode_kelompok, b.kode_jenis, b.kode_objek, b.kode_rincian_objek, b.kode_sub_rincian_objek) = r.kode_sub_rincian_objek AND b.kode_sub_sub_rincian_objek IS NULL
				END
			WHERE
				(r.pidopd::text = p.pidopd OR trim(both FROM p.pidopd) = '') AND
				(r.pidopd_cabang::text = p.pidopd_cabang OR trim(both FROM p.pidopd_cabang) = '') AND
				(r.pidupt::text = p.pidupt OR trim(both FROM p.pidupt) = '') AND
				to_char(r.tgl_dibukukan, 'yyyy-mm') <= p.tanggal AND
				CASE
					WHEN p.kode_rincian_objek != '' THEN r.kode_rincian_objek = '1.3.' || p.kode_jenis || '.' || p.kode_objek || '.' || p.kode_rincian_objek
					WHEN p.kode_objek != '' THEN r.kode_objek = '1.3.' || p.kode_jenis || '.' || p.kode_objek
					WHEN p.kode_jenis != '' THEN r.kode_jenis = '1.3.' || p.kode_jenis
					ELSE TRUE
				END
		)
		SELECT
			nama_barang,
			kode_barang,
			SUM(vol_awal) AS vol_awal,
			SUM(saldo_awal_nilaiperolehan) AS saldo_awal_nilaiperolehan,
			SUM(saldo_awal_atribusi) AS saldo_awal_atribusi,
			SUM(saldo_awal_nilaiperolehan) + SUM(saldo_awal_atribusi) AS saldo_awal_perolehanatribusi,
			SUM(vol_tambah) AS vol_tambah,
			SUM(mutasi_tambah_nilaiperolehan) AS mutasi_tambah_nilaiperolehan,
			SUM(mutasi_tambah_atribusi) AS mutasi_tambah_atribusi,
			SUM(mutasi_tambah_nilaiperolehan) + SUM(mutasi_tambah_atribusi) AS mutasi_tambah_perolehanatribusi,
			SUM(vol_kurang) AS vol_kurang,
			SUM(mutasi_kurang_nilaiperolehan) AS mutasi_kurang_nilaiperolehan,
			SUM(mutasi_kurang_atribusi) AS mutasi_kurang_atribusi,
			SUM(mutasi_kurang_nilaiperolehan) + SUM(mutasi_kurang_atribusi) AS mutasi_kurang_perolehanatribusi,
			SUM(vol_awal) + SUM(vol_tambah) - SUM(vol_kurang) AS vol_akhir,
			SUM(saldo_awal_nilaiperolehan) + SUM(mutasi_tambah_nilaiperolehan) - SUM(mutasi_kurang_nilaiperolehan) AS saldo_akhir_nilaiperolehan,
			SUM(saldo_awal_atribusi) + SUM(mutasi_tambah_atribusi) - SUM(mutasi_kurang_atribusi) AS saldo_akhir_atribusi,
			(SUM(saldo_awal_nilaiperolehan) + SUM(saldo_awal_atribusi)) + (SUM(mutasi_tambah_nilaiperolehan) + SUM(mutasi_tambah_atribusi)) - (SUM(mutasi_kurang_nilaiperolehan) + SUM(mutasi_kurang_atribusi)) AS saldo_akhir_perolehanatribusi
		FROM list
		GROUP BY 1, 2
		ORDER BY 2
	`

	// main query
	if params.Export == "export" {
		// if err := sqlQuery.Offset(start).Limit(limit).Find(&report).Error; err != nil {
		err := i.db.Raw(query, tahun_sk, tahun_sb, params.Tgl, params.Pidopd, params.PidopdCabang, params.Pidupt, params.Jenisrekap, params.KodeJenis, params.KodeObjek, params.KodeRincianObjek).Scan(&inventaris).Error
		if err != nil {
			log.Println("error", err.Error())
		}

		return inventaris, 0, 0, 1, 0, nil
	}

	query = query + fmt.Sprintf(" LIMIT %v OFFSET %v ", params.Limit, params.Start)
	err := i.db.Raw(query, tahun_sk, tahun_sb, params.Tgl, params.Pidopd, params.PidopdCabang, params.Pidupt, params.Jenisrekap, params.KodeJenis, params.KodeObjek, params.KodeRincianObjek).Scan(&inventaris).Error
	if err != nil {
		log.Println("error", err.Error())
	}

	var countData int64

	// get count filtered
	strWhere := fmt.Sprintf(`
			(r.pidopd::text ='%s' OR trim(both from '%s')='') and
			(r.pidopd_cabang::text ='%s' OR trim(both from '%s')='') and
			(r.pidupt::text ='%s' OR trim(both from '%s')='')
			and to_char(r.tgl_dibukukan, 'yyyy-mm') <= '%s' `, params.Pidopd, params.Pidopd, params.PidopdCabang, params.PidopdCabang, params.Pidupt, params.Pidupt, params.Tgl)
	// get from cache
	err = i.redisCache.Get(context.TODO(), "mutasibmd-count"+strWhere+params.Jenisrekap, &countData)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, 1, 0, err
	}

	sqlCount := fmt.Sprintf(`with params as (
                select ?::text tanggal, ?::text pidopd, ?::text pidopd_cabang, ?::text pidupt, ?::text jenis
            ), total as (
				select b.nama_rek_aset,
				case when p.jenis = '1' then r.kode_jenis
						when p.jenis ='2' then r.kode_objek
						when p.jenis ='3' then r.kode_rincian_objek
						when p.jenis ='4' then r.kode_sub_rincian_objek end kode_barang
				from reportrekap r
				cross join params as p
				inner join m_barang as b on
					case when p.jenis = '1' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis)=r.kode_jenis and b.kode_objek is null and b.kode_jenis is not null
						when p.jenis='2' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek)=r.kode_objek and b.kode_rincian_objek is null
						when p.jenis ='3' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek)=r.kode_rincian_objek and b.kode_sub_rincian_objek is null
						when p.jenis ='4' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek,b.kode_sub_rincian_objek)=r.kode_sub_rincian_objek and b.kode_sub_sub_rincian_objek is null
						end
				where %s group by 1,2
			) select count(nama_rek_aset) from total;`, strWhere)

	if err == cache.ErrCacheMiss || countData == 0 {
		sqlTxCountFiltered := i.db.Raw(sqlCount, params.Tgl, params.Pidopd, params.PidopdCabang, params.Pidupt, params.Jenisrekap).Count(&countData)

		if sqlTxCountFiltered.Error != nil {
			return nil, 0, 0, 1, 0, sqlTxCountFiltered.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   "mutasibmd-count" + strWhere + params.Jenisrekap,
			Value: countData,
			TTL:   time.Minute * 10,
		})
	}

	// var countDataFiltered int64
	countDataFiltered := countData

	var ndraw int64
	if params.Draw != "" {
		ndraw, _ = strconv.ParseInt(params.Draw, 10, 64)
	}

	summary_perpage := models.SummaryMutasi{}
	for _, np := range inventaris {
		summary_perpage.SaldoawalNilaiperolehan = summary_perpage.SaldoawalNilaiperolehan + np.SaldoAwalNilaiperolehan
		summary_perpage.MutasitambahNilaiperolehan = summary_perpage.MutasitambahNilaiperolehan + np.MutasiTambahNilaiperolehan
		summary_perpage.MutasikurangNilaiperolehan = summary_perpage.MutasikurangNilaiperolehan + np.MutasiKurangNilaiperolehan
		summary_perpage.SaldoakhirNilaiperolehan = summary_perpage.SaldoakhirNilaiperolehan + np.SaldoAkhirNilaiperolehan

		summary_perpage.SaldoawalAtribusi = summary_perpage.SaldoawalAtribusi + np.SaldoAwalAtribusi
		summary_perpage.MutasitambahAtribusi = summary_perpage.MutasitambahAtribusi + np.MutasiTambahAtribusi
		summary_perpage.MutasikurangAtribusi = summary_perpage.MutasikurangAtribusi + np.MutasiKurangAtribusi
		summary_perpage.SaldoakhirAtribusi = summary_perpage.SaldoakhirAtribusi + np.SaldoAkhirAtribusi

		summary_perpage.SaldoawalPerolehanatribusi = summary_perpage.SaldoawalPerolehanatribusi + np.SaldoAwalPerolehanatribusi
		summary_perpage.MutasitambahPerolehanatribusi = summary_perpage.MutasitambahPerolehanatribusi + np.MutasiTambahPerolehanatribusi
		summary_perpage.MutasikurangPerolehanatribusi = summary_perpage.MutasikurangPerolehanatribusi + np.MutasiKurangPerolehanatribusi
		summary_perpage.SaldoakhirPerolehanatribusi = summary_perpage.SaldoakhirPerolehanatribusi + np.SaldoAkhirPerolehanatribusi
	}

	return inventaris, countData, countDataFiltered, ndraw, summary_perpage, nil
}

func (i *reportMutasiBMDUseCase) GetTotal(start int, limit int, g *gin.Context) (*models.SummaryMutasi, error) {
	tgl := ""
	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""
	// tglakhir := ""

	if g.Query("f_jenisperiode") == "1" { // semua
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
	} else { // range
		if g.Query("f_periode") == "1" { // bulan
			tgl = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-01"
			// tglakhir = g.Query("f_tahun") + "-" + g.Query("f_bulan") + "-31"
		} else if g.Query("f_periode") == "2" { // triwulan 1 (1 januari - 31 maret)
			tgl = g.Query("f_tahun") + "-01-01"
			// tglakhir = g.Query("f_tahun") + "-03-31"
		} else if g.Query("f_periode") == "3" { // semester 1 (1 januari - 31 juni)
			tgl = g.Query("f_tahun") + "-01-01"
			// tglakhir = g.Query("f_tahun") + "-06-30"
		} else if g.Query("f_periode") == "4" { // triwulan 3 (1 juli - 30 september)
			tgl = g.Query("f_tahun") + "-07-01"
			// tglakhir = g.Query("f_tahun") + "-09-30"
		} else if g.Query("f_periode") == "5" { // tahun  (1 januari - 31 desember)
			tgl = g.Query("f_tahun") + "-01-01"
			// tglakhir = g.Query("f_tahun") + "-12-31"
		}
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
		pidupt = g.Query("f_subkuasa_filter")
	} else {
		if g.Query("subkuasa_filter") != "" {
			pidupt = g.Query("subkuasa_filter")
		}
	}

	tahun_sk, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb, _ := strconv.Atoi(g.Query("f_tahun"))
	tahun_sb = tahun_sb - 1
	jenisrekap := g.Query("f_jenisrekap")
	// jenisperiode := g.Query("f_jenisperiode")
	kode_jenis := g.Query("f_jenisbarangs_filter")
	kode_objek := g.Query("f_kodeobjek_filter")
	kode_rincian_objek := g.Query("f_koderincianobjek_filter")

	summary_page := models.SummaryMutasi{}

	// pre query
	query := `with params as (
			select ?::int tahun_sekarang, ?::int tahun_sebelum, ?::text tanggal,
			?::text pidopd, ?::text pidopd_cabang, ?::text pidupt, ?::text jenis,
			?::text kode_jenis, ?::text kode_objek , ?::text kode_rincian_objek
		), list as (
			select
			b.nama_rek_aset nama_barang,
			case when p.jenis = '1' then r.kode_jenis
				when p.jenis ='2' then r.kode_objek
				when p.jenis ='3' then r.kode_rincian_objek
				when p.jenis ='4' then r.kode_sub_rincian_objek end kode_barang,
			case when to_char(r.tgl_dibukukan,'yyyy')::int < p.tahun_sekarang then r.jumlah else 0 end vol_awal,
			case when to_char(r.tgl_dibukukan,'yyyy')::int < p.tahun_sekarang then r.jumlah * r.harga_satuan else 0 end saldo_awal_nilaiperolehan,
			coalesce(pm.saldo_awal_atribusi,0) saldo_awal_atribusi,
			-- case when to_char(pm.tgl,'yyyy-mm') < p.tanggal then pm.biaya else 0 end saldo_awal_atribusi,
			case when to_char(r.tgl_dibukukan,'yyyy')::int >= p.tahun_sekarang then r.jumlah else 0 end vol_tambah,
			case when to_char(r.tgl_dibukukan,'yyyy')::int >= p.tahun_sekarang then r.jumlah * r.harga_satuan else 0 end mutasi_tambah_nilaiperolehan,
			coalesce(pm.mutasi_tambah_atribusi,0) mutasi_tambah_atribusi,
			-- case when to_char(pm.tgl,'yyyy-mm') >= p.tanggal then pm.biaya else 0 end mutasi_tambah_atribusi,
			0 vol_kurang, 0 mutasi_kurang_nilaiperolehan,0 mutasi_kurang_atribusi,
			0 vol_akhir, 0 saldo_akhir_nilaiperolehan, 0 saldo_akhir_atribusi
			from reportrekap r
			cross join params p
			left join view_pemeliharaan_02 pm on pm.pidinventaris = r.id
			-- left join pemeliharaan pm on pm.pidinventaris=r.id and to_char(pm.tgl, 'yyyy-mm') <= p.tanggal
			inner join m_barang as b on
				case when p.jenis = '1' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis)=r.kode_jenis and b.kode_objek is null and b.kode_jenis is not null
				when p.jenis='2' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek)=r.kode_objek and b.kode_rincian_objek is null
				when p.jenis ='3' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek)=r.kode_rincian_objek and b.kode_sub_rincian_objek is null
				when p.jenis ='4' then concat_ws('.',b.kode_akun,b.kode_kelompok,b.kode_jenis,b.kode_objek,b.kode_rincian_objek,b.kode_sub_rincian_objek)=r.kode_sub_rincian_objek and b.kode_sub_sub_rincian_objek is null
				end
			where
				(r.pidopd::text =p.pidopd OR trim(both from p.pidopd)='') and
				(r.pidopd_cabang::text =p.pidopd_cabang OR trim(both from p.pidopd_cabang)='') and
				(r.pidupt::text =p.pidupt OR trim(both from p.pidupt)='')
				and to_char(r.tgl_dibukukan, 'yyyy-mm') <= p.tanggal
				and case when p.kode_rincian_objek != '' then r.kode_rincian_objek ='1.3.'||p.kode_jenis||'.'||p.kode_objek||'.'||p.kode_rincian_objek
					when p.kode_objek != '' then r.kode_objek ='1.3.'||p.kode_jenis||'.'||p.kode_objek
					when p.kode_jenis != '' then r.kode_jenis = '1.3.'||p.kode_jenis
					else true end
		) select
		sum(vol_awal) vol_awal,
		sum(saldo_awal_nilaiperolehan) saldoawal_nilaiperolehan, 
		sum(saldo_awal_atribusi) saldoawal_atribusi,
		sum(saldo_awal_nilaiperolehan) + sum(saldo_awal_atribusi) saldoawal_perolehanatribusi,
		sum(vol_tambah) vol_tambah, 
		sum(mutasi_tambah_nilaiperolehan) mutasitambah_nilaiperolehan, 
		sum(mutasi_tambah_atribusi) mutasitambah_atribusi,
		sum(mutasi_tambah_nilaiperolehan) + sum(mutasi_tambah_atribusi) mutasitambah_perolehanatribusi,
		sum(vol_kurang) vol_kurang, 
		sum(mutasi_kurang_nilaiperolehan) mutasikurang_nilaiperolehan, 
		sum(mutasi_kurang_atribusi) mutasikurang_atribusi,
		sum(mutasi_kurang_nilaiperolehan) + sum(mutasi_kurang_atribusi) mutasikurang_perolehanatribusi,
		sum(vol_awal) + sum(vol_tambah) - sum(vol_kurang) vol_akhir,
		sum(saldo_awal_nilaiperolehan) + sum(mutasi_tambah_nilaiperolehan) - sum(mutasi_kurang_nilaiperolehan) saldoakhir_nilaiperolehan,
		sum(saldo_awal_atribusi) +  sum(mutasi_tambah_atribusi) - sum(mutasi_kurang_atribusi) saldoakhir_atribusi,
		(sum(saldo_awal_nilaiperolehan) + sum(saldo_awal_atribusi)) + (sum(mutasi_tambah_nilaiperolehan) + sum(mutasi_tambah_atribusi)) - (sum(mutasi_kurang_nilaiperolehan) + sum(mutasi_kurang_atribusi)) saldoakhir_perolehanatribusi
		from list `

	// get count filtered
	strWhere := fmt.Sprintf(`
		(r.pidopd::TEXT = '%s' OR TRIM(BOTH FROM '%s' ) = '') 
		AND (r.pidopd_cabang::TEXT = '%s' OR TRIM(BOTH FROM '%s') = '') 
		AND (r.pidupt::TEXT = '%s' OR TRIM(BOTH FROM '%s') = '') 
		and to_char(r.tgl_dibukukan, 'yyyy-mm') <= '%s'`, pidopd, pidopd, pidopd_cabang, pidopd_cabang, pidupt, pidupt, tgl)

	// get from cache
	err := i.redisCache.Get(context.TODO(), "mutasibmd-total"+strWhere+jenisrekap, &summary_page)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, err
	}

	if err == cache.ErrCacheMiss || summary_page.SaldoawalNilaiperolehan == 0 {
		if err := i.db.Raw(query, tahun_sk, tahun_sb, tgl, pidopd, pidopd_cabang, pidupt, jenisrekap, kode_jenis, kode_objek, kode_rincian_objek).Scan(&summary_page).Error; err != nil {
			return nil, err
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   "mutasibmd-total" + strWhere + jenisrekap,
			Value: summary_page,
			TTL:   time.Minute * 10,
		})
	}

	return &summary_page, nil
}
