package helios

import (
	"database/sql"
	"errors"
	"github.com/RangelReale/osin"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
)

type Helios struct {
	server *osin.Server
	dbmap  *gorp.DbMap
}

func NewHelios() *Helios {
	return new(Helios)
}

func (helios *Helios) tokenHandler(w http.ResponseWriter, r *http.Request) {
	resp := helios.server.NewResponse()
	defer resp.Close()
	if ar := helios.server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.PASSWORD:
			user := models.User{Name: ar.Username}
			user.FindByName(helios.dbmap)
			if user.IsValidPassword(ar.Password) {
				ar.Authorized = true
			}
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		}
		helios.server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		log.Printf("ERROR: %s\n", resp.InternalError)
	}
	osin.OutputJSON(resp, w, r)
}

func (helios *Helios) initDb(dataSourceName string, dbConfig *config.DbConfig) {
	db, err := sql.Open(dbConfig.Type, dataSourceName)
	if err != nil {
		panic(err)
	}
	helios.dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{dbConfig.Engine, dbConfig.Encoding}}
	helios.dbmap.AddTableWithName(models.User{}, dbConfig.UserTable).SetKeys(true, dbConfig.UserTableKey)
}

func (helios *Helios) initServer(redisConfig *config.RedisConfig) {
	osinConfig := osin.NewServerConfig()
	osinConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	osinConfig.AllowGetAccessRequest = true
	osinConfig.AllowClientSecretInParams = true

	redisStorage := storage.NewRedisStorage(redisConfig)
	helios.server = osin.NewServer(osinConfig, redisStorage)
}

func (helios *Helios) Run() {
	if len(os.Args) < 2 {
		log.Println("Provide mysql data source, like: user:pass@tcp(host:port)/dbname")
		panic(errors.New("No data source provided"))
	}
	dataSourceName := os.Args[1]

	conf := config.LoadConfig("./config/config.json")

	helios.initDb(dataSourceName, conf.Db)
	defer helios.dbmap.Db.Close()

	helios.initServer(conf.Redis)

	http.HandleFunc("/token", helios.tokenHandler)

	err := http.ListenAndServe(conf.Server.Address, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
