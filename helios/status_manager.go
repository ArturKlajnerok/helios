package helios

import (
	"os"
	"time"

	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/storage"
)

const (
	HardBlockFileName      = "/etc/disabled/helios"
	RedisPingInterval      = 1 //in seconds
	HardBlockCheckInterval = 1 //in seconds
)

type StatusManager struct {
	run                           bool
	isHardBlocked                 bool
	isTokenStorageMasterAvailable bool
	isTokenStorageSlaveAvailable  bool
	isUserStorageMasterAvailable  bool
	isUserStorageSlaveAvailable   bool
	forceReadOnly                 bool
}

func (statusManager *StatusManager) checkIsHardBlocked() bool {
	isHardBlocked := false
	_, err := os.Stat(HardBlockFileName)
	if err == nil {
		isHardBlocked = true
	} else if !os.IsNotExist(err) {
		panic(err)
	}

	return isHardBlocked
}

func (statusManager *StatusManager) AllowTraffic() bool {
	if statusManager.isHardBlocked {
		return false
	}

	return (statusManager.isTokenStorageMasterAvailable || statusManager.isTokenStorageSlaveAvailable) &&
		(statusManager.isUserStorageMasterAvailable || statusManager.isUserStorageSlaveAvailable)
}

func (statusManager *StatusManager) IsServiceReadOnly() bool {
	if !statusManager.AllowTraffic() {
		return false
	}

	if (!statusManager.isTokenStorageMasterAvailable && statusManager.isTokenStorageSlaveAvailable) ||
		(!statusManager.isUserStorageMasterAvailable && statusManager.isUserStorageSlaveAvailable) {
		return true
	}

	if statusManager.forceReadOnly &&
		(statusManager.isTokenStorageMasterAvailable || statusManager.isTokenStorageSlaveAvailable) &&
		(statusManager.isUserStorageMasterAvailable || statusManager.isUserStorageSlaveAvailable) {
		return true
	}

	return false
}

func (statusManager *StatusManager) runIsHardBlockedCheck() {
	for statusManager.run {
		time.Sleep(HardBlockCheckInterval * time.Second)
		statusManager.isHardBlocked = statusManager.checkIsHardBlocked()
	}
}

func (statusManager *StatusManager) runRedisMasterPing(redisStorage *storage.RedisStorage) {
	for statusManager.run {
		time.Sleep(RedisPingInterval * time.Second)
		statusManager.isTokenStorageMasterAvailable = redisStorage.PingMaster() == nil
		if !statusManager.forceReadOnly {
			redisStorage.SetForceUseSlave(!statusManager.isTokenStorageMasterAvailable)
		}
	}
}

func (statusManager *StatusManager) runRedisSlavePing(redisStorage *storage.RedisStorage) {
	for statusManager.run {
		time.Sleep(RedisPingInterval * time.Second)
		statusManager.isTokenStorageSlaveAvailable = redisStorage.PingSlave() == nil
	}
}

func NewStatusManager(serverConfig *config.ServerConfig, redisStorage *storage.RedisStorage) *StatusManager {
	statusManager := new(StatusManager)

	statusManager.run = true
	statusManager.isHardBlocked = statusManager.checkIsHardBlocked()
	statusManager.forceReadOnly = serverConfig.ForceReadOnly
	if statusManager.forceReadOnly {
		redisStorage.SetForceUseSlave(true)
	}

	go statusManager.runIsHardBlockedCheck()
	go statusManager.runRedisMasterPing(redisStorage)
	go statusManager.runRedisSlavePing(redisStorage)

	return statusManager
}

func (statusManager *StatusManager) Close() {
	statusManager.run = false
}
