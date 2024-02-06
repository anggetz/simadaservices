package usecase

import (
	"fmt"
	"simadaservices/pkg/models"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

func getLoggedInOrganisasi(tokenInfo jwt.MapClaims, dbgorm *gorm.DB) (*models.Organisasi, error) {

	organisasiLoggedIn := models.Organisasi{}

	// get organisasi
	sqlOrgTx := dbgorm.Find(&organisasiLoggedIn, fmt.Sprintf("id = %v", tokenInfo["org_id"]))
	if sqlOrgTx.Error != nil {
		return nil, sqlOrgTx.Error
	}

	return &organisasiLoggedIn, nil
}

func buildInventarisAktifWhereClauseString() []string {
	whereClauseAccess := []string{
		fmt.Sprintf("inventaris.posting_flag IS TRUE"),
		fmt.Sprintf("inventaris.deleted_at IS NULL"),
		fmt.Sprintf("inventaris.verifikator_flag IS TRUE"),
		fmt.Sprintf("inventaris.draft IS NULL"),
	}

	return whereClauseAccess
}

func buildInventarisWhereClauseString(sql *gorm.DB, dbgorm *gorm.DB, organisasiLoggedIn *models.Organisasi) (*gorm.DB, []string) {

	whereClauseAccess := []string{}

	sql = sql.Joins("left join m_organisasi as organisasi_pengguna ON organisasi_pengguna.id = inventaris.pidopd").
		Joins("left join m_organisasi as organisasi_kuasa_pengguna ON organisasi_kuasa_pengguna.id = inventaris.pidopd_cabang").
		Joins("left join m_organisasi as organisasi_sub_kuasa_pengguna ON organisasi_sub_kuasa_pengguna.id = inventaris.pidupt")

	if organisasiLoggedIn.Level == 0 {
		idsOrg := []int{}

		// get the children
		level1Orgs := []models.Organisasi{}

		sqlOrgLevel1 := dbgorm.Find(&level1Orgs, fmt.Sprintf("pid = %v", organisasiLoggedIn.ID))
		if sqlOrgLevel1.Error != nil {
			return sql, whereClauseAccess
		}

		for _, org := range level1Orgs {
			level2Orgs := []models.Organisasi{}
			sqlOrgLevel2 := dbgorm.Find(&level2Orgs, fmt.Sprintf("pid = %v", org.ID))
			if sqlOrgLevel2.Error != nil {
				return sql, whereClauseAccess
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
	return sql, whereClauseAccess
}
