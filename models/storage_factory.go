package models

import (
	"database/sql"

	"github.com/Wikia/helios/config"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

type StorageFactory struct {
	dbmapMaster   *gorp.DbMap
	dbmapSlave    *gorp.DbMap
	userStorage   *UserStorage
	storagePinger *StoragePinger
}

func (storageFactory *StorageFactory) Close() {
	storageFactory.dbmapMaster.Db.Close()
	storageFactory.dbmapSlave.Db.Close()
}

func NewStorageFactory(dbConfig *config.DbConfig) *StorageFactory {

	storageFactory := new(StorageFactory)
	dbMaster, err := sql.Open(dbConfig.Type, dbConfig.ConnectionStringMaster)
	if err != nil {
		panic(err)
	}
	dbSlave, err := sql.Open(dbConfig.Type, dbConfig.ConnectionStringSlave)
	if err != nil {
		panic(err)
	}
	storageFactory.dbmapMaster = &gorp.DbMap{Db: dbMaster, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	storageFactory.dbmapSlave = &gorp.DbMap{Db: dbSlave, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	storageFactory.dbmapMaster.AddTableWithName(User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)
	storageFactory.dbmapSlave.AddTableWithName(User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)

	storageFactory.userStorage = NewUserStorage(storageFactory.dbmapMaster, storageFactory.dbmapSlave)
	storageFactory.storagePinger = NewStoragePinger(storageFactory.dbmapMaster, storageFactory.dbmapSlave)

	return storageFactory
}

func (storageFactory *StorageFactory) GetUserStorage() *UserStorage {
	return storageFactory.userStorage
}

func (storageFactory *StorageFactory) GetStoragePinger() *StoragePinger {
	return storageFactory.storagePinger
}
