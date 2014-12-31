package helios

import (
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/go-commons/perfmonitoring"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
)

type Helios struct {
	server     *osin.Server
	controller *Controller
}

func NewHelios() *Helios {
	return new(Helios)
}

func (helios *Helios) initServer(redisStorage *storage.RedisStorage, serverConfig *config.ServerConfig) {
	osinConfig := osin.NewServerConfig()
	osinConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	osinConfig.AllowGetAccessRequest = true
	osinConfig.AllowClientSecretInParams = true
	osinConfig.AccessExpiration = int32(serverConfig.TokenExpirationInSec)

	helios.server = osin.NewServer(osinConfig, redisStorage)
}

func (helios *Helios) Run() {

	conf := config.LoadConfig("./config/config.ini")
	logger.InitLogger("helios", logger.LogLevelDebug)
	logger.GetLogger().Info("Starting Helios")

	influxdbClient, err := perfmonitoring.NewInfluxdbClient()
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}

	repositoryFactory := models.NewRepositoryFactory(&conf.Db)
	redisStorage := storage.NewRedisStorage(&conf.RedisGeneral, &conf.RedisMaster, &conf.RedisSlave, &conf.Server)
	statusManager := NewStatusManager(&conf.Server, redisStorage)

	defer statusManager.Close()
	defer redisStorage.DoClose()
	defer repositoryFactory.Close()

	helios.initServer(redisStorage, &conf.Server)

	helios.controller = NewController(influxdbClient, helios.server, repositoryFactory, redisStorage, &conf.Server)

	err = http.ListenAndServe(conf.Server.Address, nil)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}
}
