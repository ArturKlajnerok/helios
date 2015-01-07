package models

import (
	"database/sql"

	"github.com/Wikia/go-commons/logger"
	"github.com/coopernurse/gorp"
)

type UserStorage struct {
	dbmap *gorp.DbMap
}

func NewUserStorage(dbmap *gorp.DbMap) *UserStorage {
	userStorage := UserStorage{dbmap: dbmap}
	return &userStorage
}

func (userStorage *UserStorage) FindByName(userName string, mustExist bool) (*User, error) {

	user := new(User)
	err := userStorage.dbmap.SelectOne(&user, "select * from user where user_name=?", userName)
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
