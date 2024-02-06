package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"simadaservices/pkg/models"

	"gorm.io/gorm"
)

type AuthUseCase interface {
	ValidateToken(token string) (models.User, bool)
	IsUserHasAccess(userId float64, permissions []string) bool
}

type authUseCaseImpl struct {
	db *gorm.DB
}

func NewAuthUseCase(db *gorm.DB) AuthUseCase {
	return &authUseCaseImpl{
		db: db,
	}
}

func (a *authUseCaseImpl) ValidateToken(token string) (models.User, bool) {
	user := models.User{}

	sha := sha256.New()
	sha.Write([]byte(token))

	sqlTx := a.db.Find(&user, fmt.Sprintf("api_token = '%s'", hex.EncodeToString(sha.Sum(nil))))

	if sqlTx.Error != nil {
		fmt.Println("ERR", sqlTx.Error.Error())
		return user, false
	}

	return user, sqlTx.RowsAffected == 1
}

func (a *authUseCaseImpl) IsUserHasAccess(userId float64, permissions []string) bool {

	var countSpatiePerm int64
	a.db.Table("model_has_roles").
		Where("model_has_roles.model_id = ? ", userId).
		Joins("left join role_has_permissions on role_has_permissions.role_id = model_has_roles.role_id").
		Joins("inner join permissions on permissions.id = role_has_permissions.permission_id").
		Where("permissions.name IN ?", permissions).Count(&countSpatiePerm)

	var countUserGroupRoleId int64
	a.db.Table("users").
		Where("users.id = ?", userId).
		Joins("inner join user_group_role_permission ug on ug.uuid = users.user_group_role_id").
		Where("ug.module_name IN ?", permissions).
		Count(&countUserGroupRoleId)

	fmt.Println(countSpatiePerm, countUserGroupRoleId)

	return countSpatiePerm > 0 || countUserGroupRoleId > 0
}
