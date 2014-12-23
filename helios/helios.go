package helios

import (
	"database/sql"
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/go-commons/perfmonitoring"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

type Helios struct {
	server     *osin.Server
	dbmap      *gorp.DbMap
	controller *Controller
}

func NewHelios() *Helios {
	return new(Helios)
}

func (helios *Helios) initDb(dataSourceName string, dbConfig *config.DbConfig) {
	db, err := sql.Open(dbConfig.Type, dataSourceName)
	if err != nil {
		panic(err)
	}
	helios.dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	helios.dbmap.AddTableWithName(models.User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)
}

func (helios *Helios) initServer(redisConfig *config.RedisConfig, serverConfig *config.ServerConfig) {
	osinConfig := osin.NewServerConfig()
	osinConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	osinConfig.AllowGetAccessRequest = true
	osinConfig.AllowClientSecretInParams = true
	osinConfig.AccessExpiration = int32(serverConfig.TokenExpirationInSec)

	redisStorage := storage.NewRedisStorage(redisConfig, serverConfig)
	helios.server = osin.NewServer(osinConfig, redisStorage)
}

func (helios *Helios) Run(dataSourceName string) {

	conf := config.LoadConfig("./config/config.json")
	logger.InitLogger("helios", logger.LogLevelDebug)
	logger.GetLogger().Info("Starting Helios")

	influxdbClient, err := perfmonitoring.NewInfluxdbClient()
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}

	helios.initDb(dataSourceName, conf.Db)
	defer helios.dbmap.Db.Close()

	helios.initServer(conf.Redis, conf.Server)

	helios.controller = NewController(influxdbClient, helios.dbmap, helios.server)

	err = http.ListenAndServe(conf.Server.Address, nil)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}
}
