package models

import (
	"database/sql"

	"github.com/Wikia/helios/config"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

type StorageFactory struct {
	dbmap       *gorp.DbMap
	userStorage *UserStorage
}

func (storageFactory *StorageFactory) Close() {
	storageFactory.dbmap.Db.Close()
}

func NewStorageFactory(dbConfig *config.DbConfig) *StorageFactory {

	storageFactory := new(StorageFactory)
	db, err := sql.Open(dbConfig.Type, dbConfig.ConnectionString)
	if err != nil {
		panic(err)
	}
	storageFactory.dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	storageFactory.dbmap.AddTableWithName(User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)

	storageFactory.userStorage = NewUserStorage(storageFactory.dbmap)

	return storageFactory
}

func (storageFactory *StorageFactory) GetUserStorage() *UserStorage {
	return storageFactory.userStorage
}
