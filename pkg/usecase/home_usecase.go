package usecase

import (
	"simadaservices/pkg/models"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

type HomeUseCase interface {
	GetTotalAset(jwt.MapClaims) (int64, error)
	GetNilaiAsset(jwt.MapClaims) (float64, error)
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

func (hu *homeUseCaseImpl) GetTotalAset(tokenInfo jwt.MapClaims) (int64, error) {

	sql := hu.db

	modelTotal := struct {
		Total int64
	}{}

	organisasiLoggedIn, err := getLoggedInOrganisasi(tokenInfo, hu.db)
	if err != nil {
		return 0, err
	}

	whereClauseAccess := []string{}

	sql, whereClauseAccess = buildInventarisWhereClauseString(sql, hu.db, organisasiLoggedIn)

	whereClauseAktifInventaris := buildInventarisAktifWhereClauseString()

	whereClauseAccess = append(whereClauseAccess, whereClauseAktifInventaris...)

	sql.Select("COUNT(inventaris.id) as total").Model(new(models.Inventaris)).Where(strings.Join(whereClauseAccess, " AND ")).Scan(&modelTotal)

	return modelTotal.Total, nil

}

func (hu *homeUseCaseImpl) GetNilaiAsset(tokenInfo jwt.MapClaims) (float64, error) {

	sql := hu.db

	modelTotal := struct {
		Total float64
	}{}

	organisasiLoggedIn, err := getLoggedInOrganisasi(tokenInfo, hu.db)
	if err != nil {
		return 0, err
	}

	whereClauseAccess := []string{}

	sql, whereClauseAccess = buildInventarisWhereClauseString(sql, hu.db, organisasiLoggedIn)

	whereClauseAktifInventaris := buildInventarisAktifWhereClauseString()

	whereClauseAccess = append(whereClauseAccess, whereClauseAktifInventaris...)

	sql.Select("SUM(inventaris.harga_satuan) as total").Model(new(models.Inventaris)).Where(strings.Join(whereClauseAccess, " AND ")).Scan(&modelTotal)

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
	whereClauseAccess := []string{}

	sql, whereClauseAccess = buildInventarisWhereClauseString(sql, hu.db, organisasiLoggedIn)

	whereClauseAktifInventaris := buildInventarisAktifWhereClauseString()

	whereClauseAccess = append(whereClauseAccess, whereClauseAktifInventaris...)

	sql.
		Select("SUM(inventaris.harga_satuan) + COALESCE(SUM(p.biaya), 0) as nilai, COUNT(1) as total, barang.kode_jenis as jenis_asset").
		Model(new(models.Inventaris)).
		Joins("LEFT JOIN pemeliharaan p ON p.pidinventaris = inventaris.id").
		Joins("JOIN m_barang barang ON barang.id = inventaris.pidbarang").
		Group("barang.kode_jenis").
		Find(&modelTotal, strings.Join(whereClauseAccess, " AND "))

	return modelTotal, nil

}
