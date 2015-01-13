package helios

import (
	"fmt"
	"strings"
	"testing"
)

const (
	HeartbeatEndpoint         = "/heartbeat"
	HealthCheckNagiosEndpoint = "/healthcheck_nagios"
)

func TestE2eHeartbeat(t *testing.T) {
	skipInShortMode(t)

	address := ServerAddress + HeartbeatEndpoint
	responseBytes := getResponse(address, t)
	response := strings.Trim(string(responseBytes), "\n")
	if response != "Service status: OK" {
		t.Fatal(fmt.Sprintf("Invalid heartbeat response received: %s. Expected: %s", response, "Service status: OK"))
	}
}

func TestE2eNagiosHealthCheck(t *testing.T) {
	skipInShortMode(t)

	address := ServerAddress + HealthCheckNagiosEndpoint
	responseBytes := getResponse(address, t)
	response := strings.Trim(string(responseBytes), "\n")
	if response != "Service status: OK" {
		t.Fatal(fmt.Sprintf("Invalid nagios healthcheck response received %s", response))
	}
}
