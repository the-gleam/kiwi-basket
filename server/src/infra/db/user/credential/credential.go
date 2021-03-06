package credential

import (
	"fmt"

	"github.com/jinzhu/gorm"
	credentialModel "github.com/team-gleam/kiwi-basket/server/src/domain/model/user/credential"
	"github.com/team-gleam/kiwi-basket/server/src/domain/model/user/token"
	"github.com/team-gleam/kiwi-basket/server/src/domain/model/user/username"
	credentialRepository "github.com/team-gleam/kiwi-basket/server/src/domain/repository/user/credential"
	"github.com/team-gleam/kiwi-basket/server/src/infra/db/handler"
)

type CredentialRepository struct {
	dbHandler *handler.DbHandler
}

func NewCredentialRepository(h *handler.DbHandler) credentialRepository.ICredentialRepository {
	h.Db.AutoMigrate(Auth{})
	return &CredentialRepository{h}
}

type Auth struct {
	Username string `gorm:"primary_key"`
	Token    string `gorm:"primary_key"`
}

func toRecord(a credentialModel.Auth) Auth {
	return Auth{a.Username().Name(), a.Token().Token()}
}

func fromRecord(a Auth) (credentialModel.Auth, error) {
	u, err := username.NewUsername(a.Username)
	return credentialModel.NewAuth(u, token.NewToken(a.Token)), err
}

func (r *CredentialRepository) Append(a credentialModel.Auth) error {
	d := toRecord(a)
	return r.dbHandler.Db.Create(&d).Error
}

func (r *CredentialRepository) Remove(u username.Username) error {
	err := r.dbHandler.Db.Where("username = ?", u.Name()).Delete(Auth{}).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return err
}

func (r *CredentialRepository) Exists(t token.Token) (bool, error) {
	a := new(Auth)
	err := r.dbHandler.Db.Where("token = ?", t.Token()).Take(a).Error
	if gorm.IsRecordNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return a.Token != "", nil
}

func (r *CredentialRepository) GetByToken(t token.Token) (credentialModel.Auth, error) {
	auth := new(Auth)
	err := r.dbHandler.Db.Where("token = ?", t.Token()).Take(auth).Error
	if err != nil {
		return credentialModel.Auth{}, err
	}

	a, err := fromRecord(*auth)
	if err != nil && err.Error() == username.InvalidUsername {
		return credentialModel.Auth{}, fmt.Errorf("user not found")
	}

	return a, nil
}

func (r *CredentialRepository) GetByUsername(u username.Username) (credentialModel.Auth, error) {
	auth := new(Auth)
	err := r.dbHandler.Db.Where("username = ?", u.Name()).Take(auth).Error
	if err != nil {
		return credentialModel.Auth{}, err
	}

	a, err := fromRecord(*auth)
	if err != nil && err.Error() == username.InvalidUsername {
		return credentialModel.Auth{}, fmt.Errorf("user not found")
	}

	return a, nil
}
