package helios

import (
	"fmt"
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
	"github.com/influxdb/influxdb/client"
)

type OAuthController struct {
	server         *osin.Server
	userStorage    *models.UserStorage
	redisStorage   *storage.RedisStorage
	influxdbClient *client.Client

	allowMultipleAccessTokens bool
}

func NewOAuthController(
	influxdbClient *client.Client,
	server *osin.Server,
	storageFactory *models.StorageFactory,
	redisStorage *storage.RedisStorage,
	serverConfig *config.ServerConfig) *OAuthController {

	controller := new(OAuthController)
	controller.influxdbClient = influxdbClient
	controller.userStorage = storageFactory.GetUserStorage()
	controller.redisStorage = redisStorage
	controller.server = server
	controller.allowMultipleAccessTokens = serverConfig.AllowMultipleAccessTokens

	http.HandleFunc("/info", controller.infoHandler)
	http.HandleFunc("/token", controller.tokenHandler)

	return controller
}

func (controller *OAuthController) infoHandler(w http.ResponseWriter, r *http.Request) {
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

func (controller *OAuthController) tokenHandlerPassword(ar *osin.AccessRequest) error {
	user, err := controller.userStorage.FindByName(ar.Username, false)
	if user != nil && user.IsValidPassword(ar.Password) {
		ar.UserData = fmt.Sprintf("%d", user.Id)
		ar.Authorized = true
		if !controller.allowMultipleAccessTokens {
			var accessData *osin.AccessData
			userId := fmt.Sprintf("%d", user.Id)
			accessData, err = controller.redisStorage.GetAccessForUserId(userId)
			if err == nil && accessData != nil {
				ar.ForceAccessData = accessData //Reuse previous token if it exists
			}
		}
	} else {
		ar.Authorized = false
		if user == nil && err == nil {
			logger.GetLogger().Debug("tokenHandlerPassword: user with the given name not found")
		} else if user != nil {
			logger.GetLogger().Debug("tokenHandlerPassword: incorrect password provided")
		}
	}

	return err
}

func (controller *OAuthController) tokenHandler(w http.ResponseWriter, r *http.Request) {
	timer := createTimerForAPICall(controller.influxdbClient, "tokenHandler")
	defer closeTimer(timer)

	resp := controller.server.NewResponse()
	defer resp.Close()

	if ar := controller.server.HandleAccessRequest(resp, r); ar != nil {
		var err error
		switch ar.Type {
		case osin.PASSWORD:
			err = controller.tokenHandlerPassword(ar)
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		}

		controller.server.FinishAccessRequest(resp, r, ar)
		if resp.InternalError != nil {
			logger.GetLogger().ErrorErr(resp.InternalError)
		} else if err == nil {
			logger.GetLogger().Debug("Successfully processed tokenHandler")
		}
	}
	osin.OutputJSON(resp, w, r)
}
