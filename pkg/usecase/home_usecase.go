package usecase

import (
	"simadaservices/pkg/models"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomeUseCase interface {
	GetTotalAset(jwt.MapClaims, string) (int64, error)
	GetNilaiAsset(jwt.MapClaims, *gin.Context) (float64, error)
	GetNilaiAssetGroupByKodeJenis(jwt.MapClaims) ([]getNilaiAssetGroupByKodeJenis, error)
}

type homeUseCaseImpl struct {
	db *gorm.DB
}

func NewHomeUseCase(db *gorm.DB) HomeUseCase {
	return &homeUseCaseImpl{
		db: db,
	}
}

func (hu *homeUseCaseImpl) GetTotalAset(tokenInfo jwt.MapClaims, kode_jenis string) (int64, error) {

	sql := hu.db

	modelTotal := struct {
		Total int64
	}{}

	organisasiLoggedIn, err := getLoggedInOrganisasi(tokenInfo, hu.db)
	if err != nil {
		return 0, err
	}

	// whereClauseAccess := []string{}

	sql, whereClauseAccess := buildInventarisWhereClauseString(sql, hu.db, organisasiLoggedIn)

	whereClauseAktifInventaris := buildInventarisAktifWhereClauseString()

	whereClauseAccess = append(whereClauseAccess, whereClauseAktifInventaris...)

	sql = sql.Select("COUNT(inventaris.id) as total").Model(new(models.Inventaris)).Where(strings.Join(whereClauseAccess, " AND ")).Joins("JOIN m_barang mb ON mb.id = inventaris.pidbarang")

	if kode_jenis != "" {
		sql = sql.Where("mb.kode_jenis = ?", kode_jenis)
	}

	sql.Scan(&modelTotal)

	return modelTotal.Total, nil

}

func (hu *homeUseCaseImpl) GetNilaiAsset(tokenInfo jwt.MapClaims, g *gin.Context) (float64, error) {

	sql := hu.db

	modelTotal := struct {
		Total float64
	}{}

	organisasiLoggedIn, err := getLoggedInOrganisasi(tokenInfo, hu.db)
	if err != nil {
		return 0, err
	}

	// whereClauseAccess := []string{}
	q := QueryParamInventaris{}
	t, _ := g.Get("token_info")

	q.Start = 0
	q.Limit = 0
	q.CanDelete = false
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

	sql, whereClauseAccess := buildInventarisWhereClauseString(sql, hu.db, organisasiLoggedIn)

	whereClauseAktifInventaris := buildInventarisAktifWhereClauseString()

	whereClauseAccess = append(whereClauseAccess, whereClauseAktifInventaris...)

	whereClause, depJoin := new(invoiceUseCaseImpl).buildGetInventarisFilter(q, true, g)

	whereClause = append(whereClause, whereClauseAccess...)

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

	if len(whereClause) > 0 {
		sql = sql.Joins("join m_barang ON m_barang.id = inventaris.pidbarang")
	}

	sql = sql.Select("SUM(inventaris.harga_satuan) + COALESCE(SUM(p.biaya), 0) as total").Model(new(models.Inventaris)).Where(strings.Join(whereClause, " AND ")).
		Joins("LEFT JOIN pemeliharaan p ON p.pidinventaris = inventaris.id")

	kode_jenis := g.Query("kode_jenis")
	if kode_jenis != "" {
		if len(whereClause) > 0 {
			sql = sql.Where("m_barang.kode_jenis = ?", kode_jenis)
		} else {
			sql = sql.Joins("JOIN m_barang mb ON mb.id = inventaris.pidbarang").
				Where("mb.kode_jenis = ?", kode_jenis)
		}
	}

	sql.Scan(&modelTotal)

	return modelTotal.Total, nil

}

type getNilaiAssetGroupByKodeJenis struct {
	Nilai      float64 `json:"nilai"`
	Total      int64   `json:"total"`
	JenisAsset string  `json:"jenis_asset"`
}

func (hu *homeUseCaseImpl) GetNilaiAssetGroupByKodeJenis(tokenInfo jwt.MapClaims) ([]getNilaiAssetGroupByKodeJenis, error) {

	sql := hu.db

	modelTotal := []getNilaiAssetGroupByKodeJenis{}
	organisasiLoggedIn, err := getLoggedInOrganisasi(tokenInfo, hu.db)
	if err != nil {
		return modelTotal, err
	}
	// whereClauseAccess := []string{}

	sql, whereClauseAccess := buildInventarisWhereClauseString(sql, hu.db, organisasiLoggedIn)

	whereClauseAktifInventaris := buildInventarisAktifWhereClauseString()

	whereClauseAccess = append(whereClauseAccess, whereClauseAktifInventaris...)

	sql.
		Select("SUM(inventaris.harga_satuan) + COALESCE(SUM(p.biaya), 0) as nilai, COUNT(distinct inventaris.id) as total, barang.kode_jenis as jenis_asset").
		Model(new(models.Inventaris)).
		Joins("LEFT JOIN pemeliharaan p ON p.pidinventaris = inventaris.id").
		Joins("JOIN m_barang barang ON barang.id = inventaris.pidbarang").
		Group("barang.kode_jenis").
		Find(&modelTotal, strings.Join(whereClauseAccess, " AND "))

	return modelTotal, nil

}
