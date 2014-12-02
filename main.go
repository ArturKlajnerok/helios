package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
)

var server *osin.Server
var dbmap *gorp.DbMap

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	resp := server.NewResponse()
	defer resp.Close()
	if ar := server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.PASSWORD:
			user := models.User{Name: ar.Username}
			user.FindByName(dbmap)
			if user.IsValidPassword(ar.Password) {
				ar.Authorized = true
			}
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logger.GetLogger().ErrorErr(resp.InternalError)
	}
	logger.GetLogger().Debug("Token generated")
	osin.OutputJSON(resp, w, r)
}

func initDb(dataSourceName string, dbConfig *config.DbConfig) *gorp.DbMap {
	db, err := sql.Open(dbConfig.Type, dataSourceName)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		panic(err)
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	dbmap.AddTableWithName(models.User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)
	return dbmap
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Provide mysql data source, like: user:pass@tcp(host:port)/dbname")
		return
	}
	logger.InitLogger("helios", logger.LOG_LEVEL_DEBUG)
	logger.GetLogger().Info("Starting Helios")
	dataSourceName := os.Args[1]

	conf := config.LoadConfig()

	dbmap = initDb(dataSourceName, conf.Db)
	defer dbmap.Db.Close()

	osinConfig := osin.NewServerConfig()
	osinConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	osinConfig.AllowGetAccessRequest = true
	osinConfig.AllowClientSecretInParams = true

	redisStorage := storage.NewRedisStorage(conf.Redis)
	server = osin.NewServer(osinConfig, redisStorage)

	http.HandleFunc("/token", tokenHandler)
	http.ListenAndServe(conf.Server.Address, nil)
}
