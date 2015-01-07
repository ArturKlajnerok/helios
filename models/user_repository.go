package models

import (
	"database/sql"

	"github.com/Wikia/go-commons/logger"
	"github.com/coopernurse/gorp"
)

type UserRepository struct {
	dbmap *gorp.DbMap
}

func NewUserRepository(dbmap *gorp.DbMap) *UserRepository {
	userRepository := UserRepository{dbmap: dbmap}
	return &userRepository
}

func (userRepository *UserRepository) FindByName(userName string, mustExist bool) (*User, error) {

	user := new(User)
	err := userRepository.dbmap.SelectOne(&user, "select * from user where user_name=?", userName)
	if err != nil {
		if err == sql.ErrNoRows && !mustExist {
			return nil, nil
		} else {
			logger.GetLogger().ErrorErr(err)
			return nil, err
		}
	}
	return user, nil
}
