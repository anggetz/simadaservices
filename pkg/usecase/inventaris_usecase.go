package usecase

import (
	"fmt"
	"simadaservices/pkg/models"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InvoiceUseCase interface {
	Get(limit, skip int, g *gin.Context) (interface{}, int64, error)
}

type invoiceUseCaseImpl struct {
	db *gorm.DB
}

func NewInventarisUseCase(db *gorm.DB) InvoiceUseCase {
	return &invoiceUseCaseImpl{
		db: db,
	}
}

type getInvoiceResponse struct {
	*models.Inventaris
	NamaRekAset    string `json:"nama_rek_aset"`
	KelompokKib    string `json:"kelompok_kib"`
	Jenis          string `json:"jenis"`
	PenggunaBarang string `json:"pengguna_barang"`
}

func (i *invoiceUseCaseImpl) Get(limit, page int, g *gin.Context) (interface{}, int64, error) {

	inventaris := []getInvoiceResponse{}

	whereClause := []string{}
	depJoin := map[string]bool{}

	// get the filter
	if g.Query("published") != "" {
		whereClause = append(whereClause, "inventaris.id_publish NOT NULL")
	}

	if g.Query("except-id-inventaris") != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.id IN (%s)", g.Query("except-id-inventaris")))
	}

	if g.Query("jenisbarangs") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_jenis = '%s'", g.Query("jenisbarangs")))
	}

	if g.Query("pencarian_khusus") != "" && g.Query("pencarian_khusus_nilai") != "" {
		switch g.Query("pencarian_khusus") {
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
					g.Query("pencarian_khusus_nilai"),
					g.Query("pencarian_khusus_nilai"),
					g.Query("pencarian_khusus_nilai"),
					g.Query("pencarian_khusus_nilai"),
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
				whereClause = append(whereClause, fmt.Sprintf("m_merk_barang.nama ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["m_merk_barang"] = true
				break
			}
		case "a.status_tanah":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.status_sertifikat ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_tanah"] = true
				break
			}
		case "a.penggunaan":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.penggunaan ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_tanah"] = true
				break
			}
		case "a.nomor_sertifikat":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.nomor_sertifikat ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_tanah"] = true
				break
			}
		case "a.status_sertifikat":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_tanah.status_sertifikat ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_tanah"] = true
				break
			}
		case "b.nomor_rangka":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_mesin.norangka ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_mesin"] = true
				break
			}
		case "b.nomor_mesin":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_mesin.nomesin ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_mesin"] = true
				break
			}
		case "b.nomor_polisi":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_mesin.nopol ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_mesin"] = true
				break
			}
		case "b.koderuasjalan":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_jalan.kode_jalan ~* '%s'", g.Query("pencarian_khusus_nilai")))
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
					g.Query("pencarian_khusus_nilai")))
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
					g.Query("pencarian_khusus_nilai"),
					g.Query("pencarian_khusus_nilai")))
				depJoin["detil_aset_lainnya"] = true
				break
			}
		case "e.jenis":
			{
				whereClause = append(whereClause, fmt.Sprintf("detil_aset_lainnya.ternak_jenis ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["detil_aset_lainnya"] = true
				break
			}
		case "inventaris.alamat_kota":
			{
				whereClause = append(whereClause, fmt.Sprintf("m_kota.nama ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["m_kota"] = true
				break
			}

		case "inventaris.alamat_kecamatan":
			{
				whereClause = append(whereClause, fmt.Sprintf("m_kecamatan.nama ~* '%s'", g.Query("pencarian_khusus_nilai")))
				depJoin["m_kecamatan"] = true
				break
			}

		default:
			{
				whereClause = append(whereClause, fmt.Sprintf("%s ~* '%s'", g.Query("pencarian_khusus"), g.Query("pencarian_khusus_nilai")))
				break
			}
		}
	}

	if g.Query("pencarian_khusus_range") != "" && g.Query("pencarian_khusus_range") != "" && (g.Query("pencarian_khusus_range_nilai_from") != "" || g.Query("pencarian_khusus_range_nilai_to") != "") {
		rangeKey := g.Query("pencarian_khusus_range")
		var from string
		if g.Query("pencarian_khusus_range_nilai_from") != "" {
			from = g.Query("pencarian_khusus_range_nilai_from")
		}

		var to string
		if g.Query("pencarian_khusus_range_nilai_to") != "" {
			to = g.Query("pencarian_khusus_range_nilai_to")
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
				fieldName = "detil_bangunan.luas"
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
						%s BETWEEN %v AND %v
					 )
					`,
				fieldName,
				from,
				to))
		} else if from != "" {
			whereClause = append(whereClause, fmt.Sprintf("%s >= %v", fieldName, from))
		} else if to != "" {
			whereClause = append(whereClause, fmt.Sprintf("%s <= %v", fieldName, to))
		}
	}

	if g.Query("kodeobjek") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kode_objek = '%s'", g.Query("kodeobjek")))
	}

	if g.Query("koderincianobjek") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.koderincianobjek = '%s'", g.Query("koderincianobjek")))
	}

	if g.Query("kodesubrincianobjek") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_barang.kodesubrincianobjek = '%s'", g.Query("kodesubrincianobjek")))
	}

	if g.Query("organisasi_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("m_organisasi.id = '%s'", g.Query("organisasi_filter")))
	}

	if g.Query("penggunafilter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd = '%s'", g.Query("penggunafilter")))
	}

	if g.Query("kuasapengguna_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidopd_cabang = '%s'", g.Query("kuasapengguna_filter")))
	}

	if g.Query("subkuasa_filter") != "" {
		whereClause = append(whereClause, fmt.Sprintf("inventaris.pidupt = '%s'", g.Query("subkuasa_filter")))
	}

	sql := i.db.
		Offset((page - 1) * limit).
		Limit(limit).
		Select([]string{
			"inventaris.*",
			"m_barang.nama_rek_aset",
			"m_jenis_barang.kelompok_kib",
			"m_jenis_barang.nama as jenis",
			"m_organisasi.nama as pengguna_barang",
		}).
		Joins("join m_barang ON m_barang.id = inventaris.pidbarang").
		Joins("join m_jenis_barang ON m_jenis_barang.kode = m_barang.kode_jenis").
		Joins("join m_organisasi ON m_organisasi.id = inventaris.pid_organisasi").
		Joins("join m_organisasi as organisasi_pengguna ON organisasi_pengguna.id = inventaris.pidopd").
		Joins("join m_organisasi as organisasi_kuasa_pengguna ON organisasi_kuasa_pengguna.id = inventaris.pidopd_cabang").
		Joins("join m_organisasi as organisasi_sub_kuasa_pengguna ON organisasi_sub_kuasa_pengguna.id = inventaris.pidupt")

	if _, ok := depJoin["detil_tanah"]; ok {
		sql.Joins("join detil_tanah ON detil_tanah.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_mesin"]; ok {
		sql.Joins("join detil_mesin ON detil_mesin.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_bangunan"]; ok {
		sql.Joins("join detil_bangunan ON detil_bangunan.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_aset_lainnya"]; ok {
		sql.Joins("join detil_aset_lainnya ON detil_aset_lainnya.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_jalan"]; ok {
		sql.Joins("join detil_jalan ON detil_jalan.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["detil_konstruksi"]; ok {
		sql.Joins("join detil_konstruksi ON detil_konstruksi.pidinventaris = inventaris.id")
	}
	if _, ok := depJoin["m_merk_barang"]; ok {
		sql.Joins("join m_merk_barang ON m_merk_barang.id = detil_mesin.merk")
	}
	// if _, ok := depJoin["penghapusan_detail"]; ok {
	// 	sql.Joins("join m_merk_barang ON m_merk_barang.id = detil_mesin.merk")
	// }
	if _, ok := depJoin["m_kota"]; ok {
		sql.Joins("join m_alamat as m_kota ON m_kota.id = inventaris.alamat_kota")
	}
	if _, ok := depJoin["m_kecamatan"]; ok {
		sql.Joins("join m_alamat as m_kecamatan ON m_kecamatan.id = inventaris.alamat_kecamatan")
	}

	organisasiLoggedIn := models.Organisasi{}

	// get organisasi
	sqlOrgTx := i.db.Find(&organisasiLoggedIn, fmt.Sprintf("id = %v", g.GetInt("user_logged_in_org_id")))
	if sqlOrgTx.Error != nil {
		return inventaris, 0, sqlOrgTx.Error
	}

	if organisasiLoggedIn.Level == 0 {
		idsOrg := []int{}

		// get the children
		level1Orgs := []models.Organisasi{}

		sqlOrgLevel1 := i.db.Find(&level1Orgs, fmt.Sprintf("pid = %v", organisasiLoggedIn.ID))
		if sqlOrgLevel1.Error != nil {
			return inventaris, 0, sqlOrgLevel1.Error
		}

		for _, org := range level1Orgs {
			level2Orgs := []models.Organisasi{}
			sqlOrgLevel2 := i.db.Find(&level2Orgs, fmt.Sprintf("pid = %v", org.ID))
			if sqlOrgLevel2.Error != nil {
				return inventaris, 0, sqlOrgLevel2.Error
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

		whereClause = append(whereClause, fmt.Sprintf(`
			organisasi_pengguna.id = %v AND
			(
				(CASE WHEN organisasi_kuasa_pengguna.id IS NULL THEN true ELSE organisasi_kuasa_pengguna.pid = %v END)
				OR
				(CASE WHEN organisasi_sub_kuasa_pengguna.id IS NULL THEN true ELSE %s

			)
		`, organisasiLoggedIn.ID, organisasiLoggedIn.ID, elseIfSubKuasaPengguna))
	} else if organisasiLoggedIn.Level == 1 {
		whereClause = append(whereClause, fmt.Sprintf(`
		( organisasi_pengguna.id = %v AND organisasi_kuasa_pengguna.id = %v )
		AND
		(CASE WHEN organisasi_sub_kuasa_pengguna.id IS NULL THEN true ELSE organisasi_sub_kuasa_pengguna.pid = %v END)
	`, organisasiLoggedIn.ID, organisasiLoggedIn.ID, organisasiLoggedIn.ID))
	} else if organisasiLoggedIn.Level == 2 {
		whereClause = append(whereClause, fmt.Sprintf(`
			(organisasi_sub_kuasa_pengguna.id = %v) 
		`, organisasiLoggedIn.ID))
	}

	// get count
	sqlCount := i.db.
		Model(new(models.Inventaris)).
		Where(strings.Join(whereClause, " AND ")).
		Joins("join m_barang ON m_barang.id = inventaris.pidbarang").
		Joins("join m_jenis_barang ON m_jenis_barang.kode = m_barang.kode_jenis").
		Joins("join m_organisasi ON m_organisasi.id = inventaris.pid_organisasi").
		Joins("join m_organisasi as organisasi_pengguna ON organisasi_pengguna.id = inventaris.pidopd").
		Joins("join m_organisasi as organisasi_kuasa_pengguna ON organisasi_kuasa_pengguna.id = inventaris.pidopd_cabang").
		Joins("join m_organisasi as organisasi_sub_kuasa_pengguna ON organisasi_sub_kuasa_pengguna.id = inventaris.pidupt")

	var countData int64

	sqlTxCount := sqlCount.Count(&countData)

	if sqlTxCount.Error != nil {
		return nil, 0, sqlCount.Error
	}

	sqlTx := sql.Find(&inventaris, strings.Join(whereClause, " AND "))

	return inventaris, countData, sqlTx.Error

}
