package models

import (
	"database/sql"
	"errors"

	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/config"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

type RepositoryFactory struct {
	dbmap          *gorp.DbMap
	userRepository *UserRepository
}

var repositoryFactory *RepositoryFactory

func Close() {
	if repositoryFactory != nil {
		repositoryFactory.dbmap.Db.Close()
		repositoryFactory = nil
	}
}

func InitRepositoryFactory(dataSourceName string, dbConfig *config.DbConfig) {

	repositoryFactory = new(RepositoryFactory)
	db, err := sql.Open(dbConfig.Type, dataSourceName)
	if err != nil {
		panic(err)
	}
	repositoryFactory.dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	repositoryFactory.dbmap.AddTableWithName(User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)

	repositoryFactory.userRepository = NewUserRepository(repositoryFactory.dbmap)
}

func ensureRepositoryFactoryWasInitialized() {
	if repositoryFactory == nil {
		err := errors.New("Repository Factory has not been initialized before accessing it")
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}
}

func GetUserRepository() *UserRepository {
	ensureRepositoryFactoryWasInitialized()
	return repositoryFactory.userRepository
}
