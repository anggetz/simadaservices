package usecase

import (
	"fmt"
	"simadaservices/pkg/models"

	"gorm.io/gorm"
)

type AuthUseCase interface {
	ValidateToken(token string) (models.User, bool)
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
	sqlTx := a.db.Find(&user, fmt.Sprintf("api_token = '%s'", user.ApiToken))

	if sqlTx.Error != nil {
		fmt.Println("ERR", sqlTx.Error.Error())
		return user, false
	}

	return user, sqlTx.RowsAffected == 1
}
