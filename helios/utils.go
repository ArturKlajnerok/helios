package helios

import (
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/go-commons/perfmonitoring"
	"github.com/influxdb/influxdb/client"
)

const (
	InfluxAppName    = "helios"
	InfluxSeriesName = "metrics"
)

func createTimerForAPICall(influxdbClient *client.Client, methodName string) *perfmonitoring.Timer {
	perfMon := perfmonitoring.NewPerfMonitoring(influxdbClient, InfluxAppName, InfluxSeriesName)
	timer := perfmonitoring.NewTimer(perfMon, "response_time")
	timer.AddValue("method_name", methodName)
	return timer
}

func closeTimer(timer *perfmonitoring.Timer) {
	err := timer.Close()
	logger.GetLogger().ErrorErr(err)
}
