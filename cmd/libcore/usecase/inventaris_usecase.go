package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"libcore/models"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/google/uuid"
	"gopkg.in/mgo.v2/bson"
	"gorm.io/gorm"
)

type InvoiceUseCase interface {
	Get(limit, skip int, canDelete bool, g *gin.Context) (interface{}, int64, int64, error)
	GetPemeliharaanInventaris(limit, skip int, g *gin.Context) (interface{}, int64, int64, error)
	GetFromElastic(limit, skip int, g *gin.Context) (interface{}, int64, int64, error)
	GetInventarisNeedVerificator(limit int, skip int, g *gin.Context) (interface{}, int64, int64, error)
	SetRegisterQueue(g *gin.Context) (*models.TaskQueue, error)
	GetExportInventaris(q QueryParamInventaris) ([]models.ReportInventaris, error)

	SetDB(*gorm.DB) InvoiceUseCase
	SetRedisCache(*cache.Cache) InvoiceUseCase
}

const (
	INVEN_EXCEL_FILE_FOLDER = "inventaris"
	INVEN_FORMAT_FILE_TIME  = "02-01-2006 15:04:05"
)

type invoiceUseCaseImpl struct {
	db         *gorm.DB
	es         *elasticsearch.Client
	redisCache *cache.Cache
}

func NewInventarisUseCase() InvoiceUseCase {
	return &invoiceUseCaseImpl{}
}

func (i *invoiceUseCaseImpl) SetDB(db *gorm.DB) InvoiceUseCase {
	i.db = db
	return i
}

func (i *invoiceUseCaseImpl) SetRedisCache(redisCache *cache.Cache) InvoiceUseCase {
	i.redisCache = redisCache
	return i
}

type getPemeliharaanInventaris struct {
	*models.Inventaris
	NamaOpd       string `json:"nama_opd"`
	NamaOpdCabang string `json:"nama_opd_cabang"`
	NamaUpt       string `json:"nama_upt"`
	NamaRekAset   string `json:"nama_rek_aset"`
	Alamat        string `json:"alamat"`
}

type QueryParamInventaris struct {
	Published                     string  `json:"published"`
	ExceptIDInventaris            string  `json:"except-id-inventaris"`
	PencarianKhusus               string  `json:"pencarian_khusus"`
	PencarianKhususNilai          string  `json:"pencarian_khusus_nilai"`
	PencarianKhususRange          string  `json:"pencarian_khusus_range"`
	PencarianKhususRangeNilaiFrom string  `json:"pencarian_khusus_range_nilai_from"`
	PencarianKhususRangeNilaiTo   string  `json:"pencarian_khusus_range_nilai_to"`
	JenisBarangs                  string  `json:"jenisbarangs"`
	KodeObjek                     string  `json:"kodeobjek"`
	KodeRincianObjek              string  `json:"koderincianobjek"`
	PenggunaFilter                string  `json:"penggunafilter"`
	KuasaPenggunaFilter           string  `json:"kuasapengguna_filter"`
	SubKuasaFilter                string  `json:"subkuasa_filter"`
	Draft                         string  `json:"draft"`
	StatusSensus                  string  `json:"status_sensus"`
	StatusVerifikasi              string  `json:"status_verifikasi"`
	Start                         int     `json:"start"`
	Limit                         int     `json:"limit"`
	CanDelete                     bool    `json:"can_delete"`
	QueueId                       int     `json:"queue_id"`
	TokenUsername                 string  `json:"token_username"`
	TokenOrg                      float64 `json:"token_org"`
	TokenId                       float64 `json:"token_id"`
}

type OpdName struct {
	Pengguna         string `default:""`
	KuasaPengguna    string `default:""`
	SubKuasaPengguna string `default:""`
}

func (i *invoiceUseCaseImpl) GetPemeliharaanInventaris(limit, start int, g *gin.Context) (interface{}, int64, int64, error) {

	resp := []getPemeliharaanInventaris{}

	organisasiLoggedIn := models.Organisasi{}

	t, _ := g.Get("token_info")

	// get organisasi
	sqlOrgTx := i.db.Find(&organisasiLoggedIn, fmt.Sprintf("id = %v", t.(jwt.MapClaims)["org_id"]))
	if sqlOrgTx.Error != nil {
		return nil, 0, 0, sqlOrgTx.Error
	}

	whereClause := []string{}

	if g.Query("f_opd_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd = %s", g.Query("f_opd_filter")))
	}

	if g.Query("f_jenis_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_jenis = '%s'", g.Query("f_jenis_filter")))
	}

	if g.Query("search[value]") != "" {
		nilai, err := strconv.Atoi(g.Query("search[value]"))
		if err != nil {
			whereClause = append(whereClause, "m_barang.nama_rek_aset like '%"+g.Query("search[value]")+"%' ")
		} else {
			whereClause = append(whereClause, fmt.Sprintf("(inventaris.harga_satuan * inventaris.jumlah) = %v", nilai))
		}
	}

	order := ""
	// order data
	if g.Query("order[0][column]") != "" {
		column := g.Query("order[0][column]")
		sort := g.Query("order[0][dir]")

		if column == "8" { // alamat
			order = fmt.Sprintf("m_kota.nama %s", sort)
		}
		if column == "7" { // tahun perolehan
			order = fmt.Sprintf("inventaris.tahun_perolehan %s", sort)
		}
		if column == "6" { // nama barang
			order = fmt.Sprintf("m_barang.nama_rek_aset %s", sort)
		}
		if column == "5" { // noreg
			order = fmt.Sprintf("inventaris.noreg %s", sort)
		}
		if column == "4" { // nilai perolehan
			order = fmt.Sprintf("(inventaris.harga_satuan * inventaris.jumlah) %s", sort)
		}
		if column == "3" { // upt
			order = fmt.Sprintf("organisasi_sub_kuasa_pengguna.nama %s", sort)
		}
		if column == "2" { // opd cabang
			order = fmt.Sprintf("organisasi_kuasa_pengguna.nama %s", sort)
		}
		if column == "1" { // opd
			order = fmt.Sprintf("organisasi_pengguna.nama %s", sort)
		}

	}

	sql := i.db.Model(new(models.Inventaris))

	whereAccessClause := []string{}

	sql, whereAccessClause = buildInventarisWhereClauseString(sql, i.db, &organisasiLoggedIn)

	whereClauseAktif := buildInventarisAktifWhereClauseString()

	whereClause = append(whereClause, whereAccessClause...)

	whereClause = append(whereClause, whereClauseAktif...)

	var countData struct {
		Total int64
	}

	sqlCount := sql.
		Where(strings.Join(whereAccessClause, " AND "))

	redisCountAllKey := "inventaris-count-all"

	if organisasiLoggedIn.Level > 0 {
		redisCountAllKey = redisCountAllKey + "-" + strconv.Itoa(organisasiLoggedIn.ID)
	}

	err := i.redisCache.Get(context.TODO(), redisCountAllKey, &countData.Total)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, fmt.Errorf("error when get redis cache: %v", err.Error())
	}

	if err == cache.ErrCacheMiss || countData.Total == 0 {
		sqlTxCount := sqlCount.Select("COUNT(1) as total").Scan(&countData)

		if sqlTxCount.Error != nil {
			return nil, 0, 0, sqlCount.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   redisCountAllKey,
			Value: countData.Total,
			TTL:   time.Minute * 10,
		})
	}

	sql = sql.
		Where(strings.Join(whereClause, " AND ")).
		Joins("join m_barang ON m_barang.id = inventaris.pidbarang")

	sqlCountFiltered := sql

	var countDataFiltered struct {
		Total int64
	}

	err = i.redisCache.Get(context.TODO(), "pemeliharaan-"+strings.Join(whereClause, " AND "), &countDataFiltered.Total)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, fmt.Errorf("error when get redis cache: %v", err.Error())
	}
	if err == cache.ErrCacheMiss || countDataFiltered.Total == 0 {
		sqlTxCountFiltered := sqlCountFiltered.Select("COUNT(1) as total").Scan(&countDataFiltered)

		if sqlTxCountFiltered.Error != nil {
			return nil, 0, 0, sqlCountFiltered.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   strings.Join(whereClause, " AND "),
			Value: countDataFiltered.Total,
			TTL:   time.Minute * 10,
		})

	}

	sql = sql.Select([]string{
		"inventaris.*",
		"organisasi_pengguna.nama as nama_opd",
		"organisasi_kuasa_pengguna.nama as nama_opd_cabang",
		"organisasi_sub_kuasa_pengguna.nama as nama_upt",
		"m_barang.nama_rek_aset",
		"m_kota.nama as alamat",
	}).Joins("left join m_alamat as m_kota ON m_kota.id = inventaris.alamat_kota")

	if order != "" {
		sql = sql.Order(order)
	}

	txData := sql.
		Offset(start).
		Limit(limit).Find(&resp)

	return &resp, countDataFiltered.Total, countData.Total, txData.Error
}

type getInvoiceResponse struct {
	*models.Inventaris
	NamaRekAset      string `json:"nama_rek_aset"`
	KelompokKib      string `json:"kelompok_kib"`
	Jenis            string `json:"jenis"`
	StatusVerifikasi string `json:"status_verifikasi"`
	PenggunaBarang   string `json:"pengguna_barang"`
	Detail           string `json:"detail"`
	CanDelete        bool   `json:"can_delete"`
}

func (i *invoiceUseCaseImpl) GetInventarisNeedVerificator(limit, start int, g *gin.Context) (interface{}, int64, int64, error) {

	inventaris := []getInvoiceResponse{}

	whereClauseAccess := []string{}
	whereClause := []string{}

	sql := i.db

	sql = sql.Model(new(models.Inventaris)).Where("inventaris.verifikator_flag IS FALSE AND inventaris.verifikator_is_revise IS FALSE").
		Joins("join m_barang ON m_barang.id = inventaris.pidbarang").
		Joins("join m_jenis_barang ON m_jenis_barang.kode = m_barang.kode_jenis").
		Joins("join m_organisasi ON m_organisasi.id = inventaris.pid_organisasi")

	organisasiLoggedIn := models.Organisasi{}

	t, _ := g.Get("token_info")

	// get organisasi
	sqlOrgTx := i.db.Find(&organisasiLoggedIn, fmt.Sprintf("id = %v", t.(jwt.MapClaims)["org_id"]))
	if sqlOrgTx.Error != nil {
		return nil, 0, 0, sqlOrgTx.Error
	}

	if organisasiLoggedIn.Level == 1 {
		whereClauseAccess = append(whereClauseAccess, "inventaris.verifikator_status = 0")
	} else if organisasiLoggedIn.Level == 0 {
		whereClauseAccess = append(whereClauseAccess, "inventaris.verifikator_status = 1")
	} else if organisasiLoggedIn.Level == 2 {
		whereClauseAccess = append(whereClauseAccess, "inventaris.verifikator_status = 10")
	}

	var countData struct {
		Total int64
	}

	sqlCount := sql.
		Model(new(models.Inventaris)).
		Where(strings.Join(whereClauseAccess, " AND "))

	redisCountAllKey := "inventaris-verificator-count-all"

	if organisasiLoggedIn.Level > 0 {
		redisCountAllKey = redisCountAllKey + "-" + strconv.Itoa(organisasiLoggedIn.ID)
	}

	err := i.redisCache.Get(context.TODO(), redisCountAllKey, &countData.Total)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, fmt.Errorf("error when get redis cache: %v", err.Error())
	}

	if err == cache.ErrCacheMiss {
		sqlTxCount := sqlCount.Select("COUNT(1) as total").Scan(&countData)

		if sqlTxCount.Error != nil {
			return nil, 0, 0, sqlCount.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   redisCountAllKey,
			Value: countData.Total,
			TTL:   time.Minute * 10,
		})
	}

	sql, whereClauseAccess = buildInventarisWhereClauseString(sql, i.db, &organisasiLoggedIn)

	whereClause = append(whereClause, whereClauseAccess...)

	sql = sql.Where(strings.Join(whereClause, " AND "))

	var countDataFiltered struct {
		Total int64
	}

	sqlCountFiltered := sql.
		Model(new(models.Inventaris)).
		Where(strings.Join(whereClause, " AND "))

	err = i.redisCache.Get(context.TODO(), "inventaris-verificator-"+strings.Join(whereClause, " AND "), &countDataFiltered.Total)
	if err != nil && err != cache.ErrCacheMiss {

		return nil, 0, 0, fmt.Errorf("error when get redis cache: %v", err.Error())
	}

	if err == cache.ErrCacheMiss || countDataFiltered.Total == 0 {
		sqlTxCountFiltered := sqlCountFiltered.Select("COUNT(1) as total").Scan(&countDataFiltered)

		if sqlTxCountFiltered.Error != nil {
			return nil, 0, 0, sqlCountFiltered.Error
		}

		err = i.redisCache.Set(&cache.Item{
			Ctx:   context.TODO(),
			Key:   strings.Join(whereClause, " AND "),
			Value: countDataFiltered.Total,
			TTL:   time.Minute * 10,
		})

	}

	sqlTx := sql.
		Select([]string{
			"inventaris.*",
			"m_barang.nama_rek_aset",
			"m_jenis_barang.kelompok_kib",
			"m_jenis_barang.nama as jenis",
			"m_organisasi.nama as pengguna_barang",
		}).
		Where(strings.Join(whereClause, " AND ")).
		Offset(start).
		Limit(limit)

	order := ""
	// order data
	if g.Query("order[0][column]") != "" {
		column := g.Query("order[0][column]")
		sort := g.Query("order[0][dir]")

		if column == "9" { // harga satuan
			order = fmt.Sprintf("inventaris.harga_satuan %s", sort)
		}
		if column == "8" { // pengguna barang
			order = fmt.Sprintf("organisasi_pengguna.nama %s", sort)
		}
		if column == "7" { // kondisi barang
			order = fmt.Sprintf("inventaris.kondisi %s", sort)
		}
		if column == "6" { // tahun perolehan
			order = fmt.Sprintf("inventaris.tahun_perolehan %s", sort)
		}
		if column == "5" { // cara perolehan
			order = fmt.Sprintf("inventaris.perolehan %s", sort)
		}
		if column == "4" { // nama barang
			order = fmt.Sprintf("m_barang.nama %s", sort)
		}
		if column == "3" { // noreg
			order = fmt.Sprintf("inventaris.noreg %s", sort)
		}
		if column == "2" { // kode barang
			order = fmt.Sprintf("inventaris.kode_barang %s", sort)
		}
	}

	if order != "" {
		sqlTx = sqlTx.Order(order)
	}
	sqlTx = sqlTx.Find(&inventaris)

	return inventaris, countDataFiltered.Total, countData.Total, sqlTx.Error
}

func (i *invoiceUseCaseImpl) buildGetInventarisFilter(q QueryParamInventaris, export bool, g *gin.Context) ([]string, map[string]bool) {
	depJoin := map[string]bool{}
	whereClause := []string{}
	// get the filter
	if q.Draft != "" {
		if q.Draft == "1" {
			whereClause = append(whereClause, "inventaris.draft IS NOT NULL")
		} else {
			whereClause = append(whereClause, "inventaris.draft IS NULL")
		}
	}

	if q.Published != "" {
		whereClause = append(whereClause, "inventaris.id_publish NOT NULL")
	}

	if q.ExceptIDInventaris != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.id IN (%s)", q.ExceptIDInventaris))
	}

	if q.PencarianKhusus != "" && q.PencarianKhususNilai != "" {
		switch q.PencarianKhusus {
		case "a,c,d,f.alamat":
			{
				whereClause = append(whereClause, fmt.Sprintf(`
						(
							detil_tanah.alamat ~* '%s' OR
							detil_bangunan.alamat ~* '%s' OR
							detil_jalan.alamat ~* '%s' OR
							detil_konstruksi.alamat ~* '%s'
						)
					`,
					q.PencarianKhususNilai,
					q.PencarianKhususNilai,
					q.PencarianKhususNilai,
					q.PencarianKhususNilai,
				))
				// depJoin = append(depJoin, "detil_tanah", "detil_bangunan", "detil_jalan", "detil_konstruksi")
				depJoin["detil_tanah"] = true
				depJoin["detil_bangunan"] = true
				depJoin["detil_jalan"] = true
				depJoin["detil_konstruksi"] = true
				break
			}
		case "b.merktipe":
			{
				whereClause = append(whereClause, fmt.Sprintf("m_merk_barang.nama ~* '%s'", q.PencarianKhususNilai))
				depJoin["m_merk_barang"] = true
				break
			}
		case "a.status_tanah":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.status_sertifikat ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_tanah"] = true
				break
			}
		case "a.penggunaan":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.penggunaan ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_tanah"] = true
				break
			}
		case "a.nomor_sertifikat":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.nomor_sertifikat ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_tanah"] = true
				break
			}
		case "a.status_sertifikat":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.status_sertifikat ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_tanah"] = true
				break
			}
		case "b.nomor_rangka":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_mesin.norangka ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_mesin"] = true
				break
			}
		case "b.nomor_mesin":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_mesin.nomesin ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_mesin"] = true
				break
			}
		case "b.nomor_polisi":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_mesin.nopol ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_mesin"] = true
				break
			}
		case "b.koderuasjalan":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_jalan.kode_jalan ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_jalan"] = true
				break
			}
		case "e.pencipta":
			{
				whereClause = append(whereClause, fmt.Sprintf(`
					 ( 
						detil_aset_lainnya.seni_pencipta ~* '%s'
					 )
					`,
					q.PencarianKhususNilai))
				depJoin["detil_aset_lainnya"] = true
				break
			}
		case "e.judulpencipta":
			{
				whereClause = append(whereClause, fmt.Sprintf(`
					 ( 
						detil_aset_lainnya.seni_pencipta ~* '%s' OR
						detil_aset_lainnya.buku_judul ~* '%s'
					 )
					`,
					q.PencarianKhususNilai,
					q.PencarianKhususNilai))
				depJoin["detil_aset_lainnya"] = true
				break
			}
		case "e.jenis":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_aset_lainnya.ternak_jenis ~* '%s'", q.PencarianKhususNilai))
				depJoin["detil_aset_lainnya"] = true
				break
			}
		case "inventaris.alamat_kota":
			{
				whereClause = append(whereClause, fmt.Sprintf("m_kota.nama ~* '%s'", q.PencarianKhususNilai))
				depJoin["m_kota"] = true
				break
			}

		case "inventaris.alamat_kecamatan":
			{
				whereClause = append(whereClause, fmt.Sprintf("m_kecamatan.nama ~* '%s'", q.PencarianKhususNilai))
				depJoin["m_kecamatan"] = true
				break
			}

		default:
			{
				whereClause = append(whereClause, fmt.Sprintf("%s ~* '%s'", q.PencarianKhusus, q.PencarianKhususNilai))
				break
			}
		}
	}

	if q.PencarianKhususRange != "" && (q.PencarianKhususRangeNilaiFrom != "" || q.PencarianKhususRangeNilaiTo != "") {
		rangeKey := q.PencarianKhususRange
		var from string
		if q.PencarianKhususRangeNilaiFrom != "" {
			from = q.PencarianKhususRangeNilaiFrom
		}

		var to string
		if q.PencarianKhususRangeNilaiTo != "" {
			to = q.PencarianKhususRangeNilaiTo
		}
		fieldName := ""
		switch rangeKey {
		case "a.luas_tanah":
			{
				fieldName = "detil_tanah.luas"
				depJoin["detil_tanah"] = true
				break
			}
		case "c.luas_bangunan":
			{
				fieldName = "detil_bangunan.luasbangunan"
				depJoin["detil_bangunan"] = true
				break
			}
		default:
			{
				fieldName = rangeKey
				break
			}
		}

		if from != "" && to != "" {
			whereClause = append(whereClause, fmt.Sprintf(`
					 ( 
						%s BETWEEN '%s' AND '%s'
					 )
					`,
				fieldName,
				from,
				to))
		} else if from != "" {
			whereClause = append(whereClause, fmt.Sprintf("%s >= '%s' ", fieldName, from))
		} else if to != "" {
			whereClause = append(whereClause, fmt.Sprintf("%s <= '%s' ", fieldName, to))
		}
	}

	if q.JenisBarangs != "" && q.JenisBarangs != "null" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_jenis = '%s'", q.JenisBarangs))
	}

	if q.KodeObjek != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_objek = '%s'", q.KodeObjek))
	}

	if q.KodeRincianObjek != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_rincian_objek = '%s'", q.KodeRincianObjek))
	}

	// if q.KodeSubRincianObjek != "" {
	// 	whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_sub_rincian_objek = '%s'", q.KodeSubRincianObjek))
	// }

	// if q.OrganisasiFilter != "" {
	// 	whereClause = append(whereClause, fmt.Sprintf("m_organisasi.id = '%s'", q.OrganisasiFilter))
	// }

	if q.PenggunaFilter != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd = '%s'", q.PenggunaFilter))
	}

	if q.KuasaPenggunaFilter != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd_cabang = '%s'", q.KuasaPenggunaFilter))
	}

	if q.SubKuasaFilter != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidupt = '%s'", q.SubKuasaFilter))
	}

	if q.StatusVerifikasi != "" {

		fmt.Println("check status verif", q.StatusVerifikasi)
		if q.StatusVerifikasi == "terverifikasi" {
			whereClause = append(whereClause, "verifikator_flag IS TRUE")
		} else if q.StatusVerifikasi == "proses verifikasi kuasa pengguna" {
			whereClause = append(whereClause, "verifikator_flag IS FALSE AND verifikator_status = 0")
		} else if q.StatusVerifikasi == "proses verifikasi pengguna" {
			whereClause = append(whereClause, "verifikator_flag IS FALSE AND verifikator_status = 1")
		} else if q.StatusVerifikasi == "revisi" {
			whereClause = append(whereClause, "verifikator_is_revise IS TRUE AND verifikator_flag IS FALSE")
		}
	}

	if !export {
		if g.Query("search[value]") != "" {
			nilai, err := strconv.Atoi(g.Query("search[value]"))
			if err != nil {
				whereClause = append(whereClause, "(m_barang.nama_rek_aset ilike '%"+g.Query("search[value]")+"%' OR inventaris.kode_barang like '%"+g.Query("search[value]")+"%' OR organisasi_pengguna.nama ilike '%"+g.Query("search[value]")+"%'"+
					"OR inventaris.noreg = '"+g.Query("search[value]")+"' OR inventaris.perolehan ilike '%"+g.Query("search[value]")+"%' OR inventaris.kondisi ilike '%"+g.Query("search[value]")+"%')")
			} else {
				whereClause = append(whereClause, fmt.Sprintf("( inventaris.harga_satuan = %v", nilai)+" OR inventaris.tahun_perolehan = '"+g.Query("search[value]")+"' )")
			}
		}
	}

	return whereClause, depJoin
}

func (i *invoiceUseCaseImpl) Get(limit, start int, canDelete bool, g *gin.Context) (interface{}, int64, int64, error) {

	inventaris := []getInvoiceResponse{}
	q := QueryParamInventaris{}

	whereClause := []string{}
	whereClauseAccess := []string{}
	depJoin := map[string]bool{}

	t, _ := g.Get("token_info")

	q.Start = start
	q.Limit = limit
	q.CanDelete = canDelete
	q.Published = g.Query("published")
	q.ExceptIDInventaris = g.Query("except-id-inventaris")
	q.PencarianKhusus = g.Query("pencarian_khusus")
	q.PencarianKhususNilai = g.Query("pencarian_khusus_nilai")
	q.PencarianKhususRange = g.Query("pencarian_khusus_range")
	q.PencarianKhususRangeNilaiFrom = g.Query("pencarian_khusus_range_nilai_from")
	q.PencarianKhususRangeNilaiTo = g.Query("pencarian_khusus_range_nilai_to")
	q.JenisBarangs = g.Query("jenisbarangs")
	q.KodeObjek = g.Query("kodeobjek")
	q.KodeRincianObjek = g.Query("koderincianobjek")
	q.PenggunaFilter = g.Query("penggunafilter")
	q.KuasaPenggunaFilter = g.Query("kuasapengguna_filter")
	q.SubKuasaFilter = g.Query("subkuasa_filter")
	q.Draft = g.Query("draft")
	q.StatusSensus = g.Query("status_sensus")
	q.StatusVerifikasi = g.Query("status_verifikasi")
	q.TokenUsername = t.(jwt.MapClaims)["username"].(string)
	q.TokenOrg = t.(jwt.MapClaims)["org_id"].(float64)
	q.TokenId = t.(jwt.MapClaims)["id"].(float64)

	// get the filter
	whereClause, depJoin = i.buildGetInventarisFilter(q, false, g)

	sql := i.db
	// Joins("join m_organisasi as organisasi_pengguna ON organisasi_pengguna.id = inventaris.pidopd").
	// Joins("join m_organisasi as organisasi_kuasa_pengguna ON organisasi_kuasa_pengguna.id = inventaris.pidopd_cabang").
	// Joins("join m_organisasi as organisasi_sub_kuasa_pengguna ON organisasi_sub_kuasa_pengguna.id = inventaris.pidupt")

	if depJoin["detil_tanah"] {
		sql = sql.Joins("join detil_tanah ON detil_tanah.pidinventaris = inventaris.id")
	}
	if depJoin["detil_mesin"] {
		sql = sql.Joins("join detil_mesin ON detil_mesin.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_bangunan"]; ok {
		sql = sql.Joins("join detil_bangunan ON detil_bangunan.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_aset_lainnya"]; ok {
		sql = sql.Joins("join detil_aset_lainnya ON detil_aset_lainnya.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_jalan"]; ok {
		sql = sql.Joins("join detil_jalan ON detil_jalan.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_konstruksi"]; ok {
		sql = sql.Joins("join detil_konstruksi ON detil_konstruksi.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["m_merk_barang"]; ok {
		sql = sql.Joins("join m_merk_barang ON m_merk_barang.id = detil_mesin.merk")
	}
	// if _, ok := depJoin["penghapusan_detail"]; ok {
	// 	sql.Joins("join m_merk_barang ON m_merk_barang.id = detil_mesin.merk")
	// }
	if _, ok := depJoin["m_kota"]; ok {
		sql = sql.Joins("join m_alamat as m_kota ON m_kota.id = inventaris.alamat_kota")
	}
	if _, ok := depJoin["m_kecamatan"]; ok {
		sql = sql.Joins("join m_alamat as m_kecamatan ON m_kecamatan.id = inventaris.alamat_kecamatan")
	}

	// get organisasi
	organisasiLoggedIn := models.Organisasi{}
	sqlOrgTx := i.db.Find(&organisasiLoggedIn, fmt.Sprintf("id = %v", q.TokenOrg))
	if sqlOrgTx.Error != nil {
		return inventaris, 0, 0, sqlOrgTx.Error
	}

	sql = sql.Joins("left join m_organisasi as organisasi_pengguna ON organisasi_pengguna.id = inventaris.pidopd").
		Joins("left join m_organisasi as organisasi_kuasa_pengguna ON organisasi_kuasa_pengguna.id = inventaris.pidopd_cabang").
		Joins(" left join m_organisasi as organisasi_sub_kuasa_pengguna ON organisasi_sub_kuasa_pengguna.id = inventaris.pidupt")

	if organisasiLoggedIn.Level == 0 {
		idsOrg := []int{}

		// get the children
		level1Orgs := []models.Organisasi{}

		sqlOrgLevel1 := i.db.Find(&level1Orgs, fmt.Sprintf("pid = %v", organisasiLoggedIn.ID))
		if sqlOrgLevel1.Error != nil {
			return inventaris, 0, 0, sqlOrgLevel1.Error
		}

		for _, org := range level1Orgs {
			level2Orgs := []models.Organisasi{}
			sqlOrgLevel2 := i.db.Find(&level2Orgs, fmt.Sprintf("pid = %v", org.ID))
			if sqlOrgLevel2.Error != nil {
				return inventaris, 0, 0, sqlOrgLevel2.Error
			}
			for _, org2 := range level1Orgs {
				idsOrg = append(idsOrg, org2.ID)
			}
			idsOrg = append(idsOrg, org.ID)
		}

		elseIfSubKuasaPengguna := "true"

		if len(idsOrg) > 0 {
			elseIfSubKuasaPengguna = fmt.Sprintf("organisasi_sub_kuasa_pengguna.id IN (%v)", strings.Trim(strings.Join(strings.Split(fmt.Sprint(idsOrg), " "), ","), "[]"))
		}

		whereClauseAccess = append(whereClauseAccess, fmt.Sprintf(`
			organisasi_pengguna.id = %v AND
			(
				(CASE WHEN organisasi_kuasa_pengguna.id IS NULL THEN true ELSE organisasi_kuasa_pengguna.pid = %v END)
				OR
				(CASE WHEN organisasi_sub_kuasa_pengguna.id IS NULL THEN true ELSE %s END)

			)
		`, organisasiLoggedIn.ID, organisasiLoggedIn.ID, elseIfSubKuasaPengguna))

	} else if organisasiLoggedIn.Level == 1 {
		whereClauseAccess = append(whereClauseAccess, fmt.Sprintf(`
		( organisasi_pengguna.id = %v AND organisasi_kuasa_pengguna.id = %v )
		AND
		(CASE WHEN organisasi_sub_kuasa_pengguna.id IS NULL THEN true ELSE organisasi_sub_kuasa_pengguna.pid = %v END)
	`, organisasiLoggedIn.ID, organisasiLoggedIn.ID, organisasiLoggedIn.ID))

	} else if organisasiLoggedIn.Level == 2 {
		whereClauseAccess = append(whereClauseAccess, fmt.Sprintf(`
			(organisasi_sub_kuasa_pengguna.id = %v) 
		`, organisasiLoggedIn.ID))
	}

	order := ""
	// order data
	if g.Query("order[0][column]") != "" {
		column := g.Query("order[0][column]")
		sort := g.Query("order[0][dir]")

		if column == "9" { // harga satuan
			order = fmt.Sprintf("inventaris.harga_satuan %s", sort)
		}
		if column == "8" { // pengguna barang
			order = fmt.Sprintf("organisasi_pengguna.nama %s", sort)
		}
		if column == "7" { // kondisi barang
			order = fmt.Sprintf("inventaris.kondisi %s", sort)
		}
		if column == "6" { // tahun perolehan
			order = fmt.Sprintf("inventaris.tahun_perolehan %s", sort)
		}
		if column == "5" { // cara perolehan
			order = fmt.Sprintf("inventaris.perolehan %s", sort)
		}
		if column == "4" { // nama barang
			order = fmt.Sprintf("m_barang.nama %s", sort)
		}
		if column == "3" { // noreg
			order = fmt.Sprintf("inventaris.noreg %s", sort)
		}
		if column == "2" { // kode barang
			order = fmt.Sprintf("inventaris.kode_barang %s", sort)
		}
	}

	sql = sql.Joins("join m_barang ON m_barang.id = inventaris.pidbarang").
		Joins("join m_jenis_barang ON m_jenis_barang.kode = m_barang.kode_jenis").
		Joins("join m_organisasi ON m_organisasi.id = inventaris.pid_organisasi")

	// get count filtered
	sqlCount := i.db.
		Table("(?) as inventaris",
			sql.Model(new(models.Inventaris)).
				Where(strings.Join(whereClauseAccess, " AND ")).
				Select("1").
				Limit(500000),
		)

	var countData struct {
		Total int64
	}

	var countDataFiltered struct {
		Total int64
	}

	sqlTxCount := sqlCount.Select("COUNT(1) as total").Scan(&countData)

	if sqlTxCount.Error != nil {
		return nil, 0, 0, sqlCount.Error
	}

	whereClause = append(whereClause, whereClauseAccess...)

	// get count filtered
	sqlCountFiltered := i.db.
		Table("(?) as inventaris",
			sql.Model(new(models.Inventaris)).
				Where(strings.Join(whereClause, " AND ")).
				Select("1").
				Limit(500000),
		)

	// get from cache
	sqlTxCountFiltered := sqlCountFiltered.Select("COUNT(1) as total").Scan(&countDataFiltered)

	if sqlTxCountFiltered.Error != nil {
		return nil, 0, 0, sqlCountFiltered.Error
	}

	sqlTx := sql.
		Select([]string{
			"inventaris.*",
			"m_barang.nama_rek_aset",
			"m_jenis_barang.kelompok_kib",
			"m_jenis_barang.nama as jenis",
			"m_organisasi.nama as pengguna_barang",
		}).
		Where(strings.Join(whereClause, " AND ")).
		Offset(q.Start).Limit(q.Limit)

	if order != "" {
		sqlTx = sqlTx.Order(order)
	}

	if err := sqlTx.Find(&inventaris).Error; err != nil {
		return nil, 0, 0, err
	}

	for _, dt := range inventaris {
		if q.CanDelete {
			dt.CanDelete = q.CanDelete
		}
		if !dt.VerifikatorFlag {
			if dt.VerifikatorIsRevise {
				dt.StatusVerifikasi = "<span class='badge bg-yellow'>Permintaan revisi data</span>"
			} else {
				if dt.VerifikatorStatus == 0 {
					dt.StatusVerifikasi = "<span class='badge bg-blue'>Proses Verifikasi Kuasa Pengguna</span>"
				} else if dt.VerifikatorStatus == 1 {
					dt.StatusVerifikasi = "<span class='badge bg-blue'>Proses Verifikasi Pengguna Barang</span>"
				} else if dt.VerifikatorStatus == 2 {
					dt.StatusVerifikasi = "<span class='badge bg-green'>Telah terverifikasi</span>"
				}
			}
		} else {
			dt.StatusVerifikasi = "<span class='badge bg-green'>Telah terverifikasi</span>"
		}
	}

	return inventaris, countDataFiltered.Total, countData.Total, sqlTx.Error
}

func (i *invoiceUseCaseImpl) GetExportInventaris(q QueryParamInventaris) ([]models.ReportInventaris, error) {
	inventaris := []models.ReportInventaris{}

	whereClause := []string{}
	whereClauseAccess := []string{}
	depJoin := map[string]bool{}

	g := &gin.Context{}

	// get the filter
	whereClause, depJoin = i.buildGetInventarisFilter(q, false, g)

	sql := i.db.Model(&[]getInvoiceResponse{})

	if depJoin["detil_tanah"] {
		sql = sql.Joins("join detil_tanah ON detil_tanah.pidinventaris = inventaris.id")
	}
	if depJoin["detil_mesin"] {
		sql = sql.Joins("join detil_mesin ON detil_mesin.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_bangunan"]; ok {
		sql = sql.Joins("join detil_bangunan ON detil_bangunan.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_aset_lainnya"]; ok {
		sql = sql.Joins("join detil_aset_lainnya ON detil_aset_lainnya.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_jalan"]; ok {
		sql = sql.Joins("join detil_jalan ON detil_jalan.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_konstruksi"]; ok {
		sql = sql.Joins("join detil_konstruksi ON detil_konstruksi.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["m_merk_barang"]; ok {
		sql = sql.Joins("join m_merk_barang ON m_merk_barang.id = detil_mesin.merk")
	}
	if _, ok := depJoin["m_kota"]; ok {
		sql = sql.Joins("join m_alamat as m_kota ON m_kota.id = inventaris.alamat_kota")
	}
	if _, ok := depJoin["m_kecamatan"]; ok {
		sql = sql.Joins("join m_alamat as m_kecamatan ON m_kecamatan.id = inventaris.alamat_kecamatan")
	}

	// get organisasi
	organisasiLoggedIn := models.Organisasi{}
	sqlOrgTx := i.db.Find(&organisasiLoggedIn, fmt.Sprintf("id = %v", q.TokenOrg))
	if sqlOrgTx.Error != nil {
		return inventaris, sqlOrgTx.Error
	}

	sql = sql.Joins("left join m_organisasi as organisasi_pengguna ON organisasi_pengguna.id = inventaris.pidopd").
		Joins("left join m_organisasi as organisasi_kuasa_pengguna ON organisasi_kuasa_pengguna.id = inventaris.pidopd_cabang").
		Joins(" left join m_organisasi as organisasi_sub_kuasa_pengguna ON organisasi_sub_kuasa_pengguna.id = inventaris.pidupt")

	if organisasiLoggedIn.Level == 0 {
		idsOrg := []int{}

		// get the children
		level1Orgs := []models.Organisasi{}

		sqlOrgLevel1 := i.db.Find(&level1Orgs, fmt.Sprintf("pid = %v", organisasiLoggedIn.ID))
		if sqlOrgLevel1.Error != nil {
			return inventaris, sqlOrgLevel1.Error
		}

		for _, org := range level1Orgs {
			level2Orgs := []models.Organisasi{}
			sqlOrgLevel2 := i.db.Find(&level2Orgs, fmt.Sprintf("pid = %v", org.ID))
			if sqlOrgLevel2.Error != nil {
				return inventaris, sqlOrgLevel2.Error
			}
			for _, org2 := range level1Orgs {
				idsOrg = append(idsOrg, org2.ID)
			}
			idsOrg = append(idsOrg, org.ID)
		}

		elseIfSubKuasaPengguna := "true"

		if len(idsOrg) > 0 {
			elseIfSubKuasaPengguna = fmt.Sprintf("organisasi_sub_kuasa_pengguna.id IN (%v)", strings.Trim(strings.Join(strings.Split(fmt.Sprint(idsOrg), " "), ","), "[]"))
		}

		whereClauseAccess = append(whereClauseAccess, fmt.Sprintf(`
			organisasi_pengguna.id = %v AND
			(
				(CASE WHEN organisasi_kuasa_pengguna.id IS NULL THEN true ELSE organisasi_kuasa_pengguna.pid = %v END)
				OR
				(CASE WHEN organisasi_sub_kuasa_pengguna.id IS NULL THEN true ELSE %s END)

			)
		`, organisasiLoggedIn.ID, organisasiLoggedIn.ID, elseIfSubKuasaPengguna))

	} else if organisasiLoggedIn.Level == 1 {
		whereClauseAccess = append(whereClauseAccess, fmt.Sprintf(`
		( organisasi_pengguna.id = %v AND organisasi_kuasa_pengguna.id = %v )
		AND
		(CASE WHEN organisasi_sub_kuasa_pengguna.id IS NULL THEN true ELSE organisasi_sub_kuasa_pengguna.pid = %v END)
	`, organisasiLoggedIn.ID, organisasiLoggedIn.ID, organisasiLoggedIn.ID))

	} else if organisasiLoggedIn.Level == 2 {
		whereClauseAccess = append(whereClauseAccess, fmt.Sprintf(`
			(organisasi_sub_kuasa_pengguna.id = %v) 
		`, organisasiLoggedIn.ID))
	}

	sql = sql.Joins("join m_barang ON m_barang.id = inventaris.pidbarang").
		Joins("join m_jenis_barang ON m_jenis_barang.kode = m_barang.kode_jenis").
		Joins("join m_organisasi ON m_organisasi.id = inventaris.pid_organisasi")

	whereClause = append(whereClause, whereClauseAccess...)

	fmt.Println(whereClause)

	sqlTx := sql.
		Select([]string{
			"inventaris.*",
			"m_barang.nama_rek_aset",
			"m_jenis_barang.kelompok_kib",
			"m_jenis_barang.nama as jenis",
			"m_organisasi.nama as pengguna_barang",
		}).
		Where(strings.Join(whereClause, " AND "))

	if err := sqlTx.Find(&inventaris).Error; err != nil {
		return nil, err
	}

	for _, dt := range inventaris {
		if !dt.VerifikatorFlag {
			if dt.VerifikatorIsRevise {
				dt.StatusVerifikasi = "Permintaan revisi data"
			} else {
				if dt.VerifikatorStatus == 0 {
					dt.StatusVerifikasi = "Proses Verifikasi Kuasa Pengguna"
				} else if dt.VerifikatorStatus == 1 {
					dt.StatusVerifikasi = "Proses Verifikasi Pengguna Barang"
				} else if dt.VerifikatorStatus == 2 {
					dt.StatusVerifikasi = "Telah terverifikasi"
				}
			}
		} else {
			dt.StatusVerifikasi = "Telah terverifikasi"
		}
	}

	return inventaris, sqlTx.Error
}

func (i *invoiceUseCaseImpl) GetFromElastic(limit, start int, g *gin.Context) (interface{}, int64, int64, error) {

	queryFilter := bson.M{}
	boolQuery := bson.M{}

	if g.Query("draft") != "" {
		if g.Query("draft") == "1" {
			boolQuery["must"] = bson.M{
				"term": bson.M{
					"draft": "1",
				},
			}
		} else {
			boolQuery["must_not"] = bson.M{
				"term": bson.M{
					"draft": "1",
				},
			}
		}
	}

	fmt.Println(queryFilter)

	query := bson.M{
		"size": 10,
		"from": start,
		"query": bson.M{
			"bool": boolQuery,
		},
	}

	byQuery, err := json.Marshal(query)
	if err != nil {
		return nil, 0, 0, err
	}

	elasticResponse := models.Elastic{}

	res, err := i.es.Search(
		i.es.Search.WithIndex("inventaris-index"),
		i.es.Search.WithBody(strings.NewReader(string(byQuery))),
	)

	if err != nil {
		return nil, 0, 0, err
	}

	err = json.NewDecoder(res.Body).Decode(&elasticResponse)

	if err != nil {
		return nil, 0, 0, err
	}

	res, err = i.es.Count(
		func(e *esapi.CountRequest) {
			e.Index = []string{"inventaris-index"}
		},
	)

	countElastic := struct {
		Count int64
	}{}

	err = json.NewDecoder(res.Body).Decode(&countElastic)

	return elasticResponse.Hits.Hits, int64(elasticResponse.Hits.Total.Value), countElastic.Count, nil
}

func (i *invoiceUseCaseImpl) SetRegisterQueue(g *gin.Context) (*models.TaskQueue, error) {

	t, _ := g.Get("token_info")
	id := t.(jwt.MapClaims)["id"].(float64)
	folderPath := os.Getenv("FOLDER_REPORT")

	// set filename
	pidopd := g.Query("penggunafilter")
	pidopd_cabang := g.Query("kuasapengguna_filter")
	pidupt := g.Query("subkuasa_filter")

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
		TaskName:     "worker-export-inventaris",
		TaskType:     "export_report",
		Status:       "pending",
		CreatedBy:    int(id),
		CallbackLink: folderPath + "/inventaris/" + fileName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := i.db.Create(&tq).Error; err != nil {
		return nil, err
	}

	return &tq, nil
}

func (i *invoiceUseCaseImpl) SetUpdateStatusQueue(g *gin.Context) (interface{}, error) {

	t, _ := g.Get("token_info")
	id := t.(jwt.MapClaims)["id"].(float64)
	folderPath := os.Getenv("FOLDER_REPORT")
	folderName := INVEN_EXCEL_FILE_FOLDER

	tq := models.TaskQueue{
		TaskUUID:     uuid.NewString(),
		TaskName:     "worker-export-inventaris",
		TaskType:     "export_report",
		Status:       "pending",
		CreatedBy:    int(id),
		CallbackLink: folderPath + "/" + folderName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := i.db.Create(&tq).Error; err != nil {
		return nil, err
	}

	return nil, nil
}

func (i *invoiceUseCaseImpl) GetOpdName(c string) OpdName {
	var params QueryParamInventaris
	opdName := OpdName{}

	err := json.Unmarshal([]byte(c), &params)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return opdName
	}

	pidopd := ""
	pidopd_cabang := ""
	pidupt := ""

	if params.PenggunaFilter != "" {
		pidopd = params.PenggunaFilter
	}

	if params.KuasaPenggunaFilter != "" {
		pidopd_cabang = params.KuasaPenggunaFilter
	}

	if params.SubKuasaFilter != "" {
		pidupt = params.SubKuasaFilter
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
