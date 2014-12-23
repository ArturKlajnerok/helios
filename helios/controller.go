package helios

import (
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/models"
	"github.com/influxdb/influxdb/client"
)

type Controller struct {
	server         *osin.Server
	userRepository *models.UserRepository
	influxdbClient *client.Client
}

func NewController(influxdbClient *client.Client, server *osin.Server) *Controller {

	controller := new(Controller)
	controller.influxdbClient = influxdbClient
	controller.userRepository = models.GetUserRepository()
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
		var err error
		switch ar.Type {
		case osin.PASSWORD:
			var user *models.User
			user, err = controller.userRepository.FindByName(ar.Username, true)
			if err == nil && user.IsValidPassword(ar.Password) {
				ar.UserData = user.Id
				ar.Authorized = true
			}
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		}

		controller.server.FinishAccessRequest(resp, r, ar)
		if resp.IsError && resp.InternalError != nil {
			logger.GetLogger().ErrorErr(resp.InternalError)
		} else if err == nil {
			logger.GetLogger().Debug("Successfully processed tokenHandler")
		}
	}
	osin.OutputJSON(resp, w, r)
}
