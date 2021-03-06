package helios

import (
	"fmt"
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/go-commons/perfmonitoring"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
)

const (
	AppName = "helios"
)

type Helios struct {
	server                *osin.Server
	oauthController       *OAuthController
	healthCheckController *HealthCheckController
}

func NewHelios() *Helios {
	return new(Helios)
}

func (helios *Helios) initServer(redisStorage *storage.RedisStorage, serverConfig *config.ServerConfig) {
	osinConfig := osin.NewServerConfig()
	osinConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	osinConfig.AllowGetAccessRequest = true
	osinConfig.AllowClientSecretInParams = true
	osinConfig.AccessExpiration = int32(serverConfig.AccessTokenExpirationInSec)

	helios.server = osin.NewServer(osinConfig, redisStorage)
}

func (helios *Helios) Run(configPath string) {

	conf := config.LoadConfig(configPath)
	logger.InitLogger(AppName, logger.LogLevelDebug)
	logger.GetLogger().Info(fmt.Sprintf("Starting %s", AppName))

	influxdbClient, err := perfmonitoring.NewInfluxdbClient()
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}

	storageFactory := models.NewStorageFactory(&conf.Db)
	redisStorage := storage.NewRedisStorage(&conf.RedisGeneral, &conf.RedisMaster, &conf.RedisSlave, &conf.Server)
	statusManager := NewStatusManager(&conf.Server, redisStorage, storageFactory)

	defer statusManager.Close()
	defer redisStorage.DoClose()
	defer storageFactory.Close()

	helios.initServer(redisStorage, &conf.Server)

	helios.oauthController = NewOAuthController(influxdbClient, helios.server, storageFactory, redisStorage, &conf.Server)
	helios.healthCheckController = NewHealthCheckController(statusManager)

	err = http.ListenAndServe(conf.Server.Address, nil)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}
}
