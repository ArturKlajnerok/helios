package main

import (
	"database/sql"
	"errors"
	"github.com/RangelReale/osin"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

const (
	DB_TYPE       = "mysql"
	DB_ENCODING   = "UTF8"
	DB_ENGINE     = "InnoDB"
	DB_KEY        = "Id"
	DB_USER_TABLE = "user"
	REDIS_HOST    = "localhost:6379"
	REDIS_PREFIX  = "auth"
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
			if user.ValidPassword(ar.Password) {
				ar.Authorized = true
			}
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		log.Printf("ERROR: %s\n", resp.InternalError)
	}
	osin.OutputJSON(resp, w, r)
}

func initDb(dataSourceName string) *gorp.DbMap {
	db, err := sql.Open(DB_TYPE, dataSourceName)
	if err != nil {
		panic(err.Error())
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{DB_ENGINE, DB_ENCODING}}
	dbmap.AddTableWithName(models.User{}, DB_USER_TABLE).SetKeys(true, DB_KEY)
	return dbmap
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Provide mysql data source, like: user:pass@tcp(host:port)/dbname")
		panic(errors.New("No data source provided"))
	}
	dataSourceName := os.Args[1]

	dbmap = initDb(dataSourceName)
	defer dbmap.Db.Close()

	cfg := osin.NewServerConfig()
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	cfg.AllowGetAccessRequest = true
	cfg.AllowClientSecretInParams = true

	server = osin.NewServer(cfg, storage.NewRedisStorage(REDIS_HOST, REDIS_PREFIX))

	http.HandleFunc("/token", tokenHandler)
	http.ListenAndServe(":8080", nil)
}
