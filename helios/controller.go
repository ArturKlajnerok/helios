package helios

import (
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/models"
	"github.com/coopernurse/gorp"
	"github.com/influxdb/influxdb/client"
)

type Controller struct {
	server         *osin.Server
	dbmap          *gorp.DbMap
	influxdbClient *client.Client
}

func NewController(influxdbClient *client.Client, dbmap *gorp.DbMap, server *osin.Server) *Controller {

	controller := new(Controller)
	controller.influxdbClient = influxdbClient
	controller.dbmap = dbmap
	controller.server = server

	http.HandleFunc("/info", controller.infoHandler)
	http.HandleFunc("/token", controller.tokenHandler)

	return controller
}

func (controller *Controller) infoHandler(w http.ResponseWriter, r *http.Request) {
	timer := createTimerForAPICall(controller.influxdbClient, "infoHandler")
	defer closeTimer(timer)

	resp := controller.server.NewResponse()
	defer resp.Close()

	if ir := controller.server.HandleInfoRequest(resp, r); ir != nil {
		controller.server.FinishInfoRequest(resp, r, ir)
		resp.Output["user_id"] = ir.AccessData.UserData
	}
	osin.OutputJSON(resp, w, r)
}

func (controller *Controller) tokenHandler(w http.ResponseWriter, r *http.Request) {
	timer := createTimerForAPICall(controller.influxdbClient, "tokenHandler")
	defer closeTimer(timer)

	resp := controller.server.NewResponse()
	defer resp.Close()

	if ar := controller.server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.PASSWORD:
			user := models.User{Name: ar.Username}
			user.FindByName(controller.dbmap)
			if user.IsValidPassword(ar.Password) {
				ar.UserData = user.Id
				ar.Authorized = true
			}
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		}
		controller.server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logger.GetLogger().ErrorErr(resp.InternalError)
	} else {
		logger.GetLogger().Debug("Successfully processed tokenHandler")
	}
	osin.OutputJSON(resp, w, r)
}
