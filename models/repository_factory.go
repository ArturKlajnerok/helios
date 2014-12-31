package models

import (
	"database/sql"

	"github.com/Wikia/helios/config"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

type RepositoryFactory struct {
	dbmap          *gorp.DbMap
	userRepository *UserRepository
}

func (repositoryFactory *RepositoryFactory) Close() {
	repositoryFactory.dbmap.Db.Close()
}

func NewRepositoryFactory(dbConfig *config.DbConfig) *RepositoryFactory {

	repositoryFactory := new(RepositoryFactory)
	db, err := sql.Open(dbConfig.Type, dbConfig.ConnectionString)
	if err != nil {
		panic(err)
	}
	repositoryFactory.dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	repositoryFactory.dbmap.AddTableWithName(User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)

	repositoryFactory.userRepository = NewUserRepository(repositoryFactory.dbmap)

	return repositoryFactory
}

func (repositoryFactory *RepositoryFactory) GetUserRepository() *UserRepository {
	return repositoryFactory.userRepository
}
