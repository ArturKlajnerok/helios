package helios

import (
	"fmt"
	"net/http"
)

type HealthCheckController struct {
	statusManager *StatusManager
}

func NewHealthCheckController(statusManager *StatusManager) *HealthCheckController {

	controller := new(HealthCheckController)
	controller.statusManager = statusManager

	http.HandleFunc("/heartbeat", controller.heartbeat)
	http.HandleFunc("/healthcheck_nagios", controller.healthCheckNagios)

	return controller
}

func (controller *HealthCheckController) heartbeat(w http.ResponseWriter, r *http.Request) {
	if controller.statusManager.AllowTraffic() {
		fmt.Fprintf(w, "Service status: OK")
	} else {
		http.Error(w, "Service status: Down", http.StatusServiceUnavailable)
	}
}

func (controller *HealthCheckController) healthCheckNagios(w http.ResponseWriter, r *http.Request) {

	status := controller.statusManager.GetStatus()

	var message string

	switch {
	case status == StatusOk:
		message = "Service status: OK"
	case status == StatusRedisMasterDown:
		message = "Service status: Redis Master Down"
	case status == StatusRedisSlaveDown:
		message = "Service status: Redis Slave Down"
	case status == StatusMySQLMasterDown:
		message = "Service status: MySQL Master Down"
	case status == StatusMySQLSlaveDown:
		message = "Service status: MySQL Slave Down"
	case status == StatusRedisAndMySQLDown:
		message = "Service status: Redis and MySQL Down"
	}

	if status == StatusOk {
		fmt.Fprintf(w, message)
	} else {
		http.Error(w, message, http.StatusServiceUnavailable)
	}
}
