package helios

import (
	"os"
	"time"

	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/config"
	"github.com/Wikia/helios/models"
	"github.com/Wikia/helios/storage"
)

const (
	HardBlockFileName      = "/etc/disabled/helios"
	RedisPingInterval      = 1  //in seconds
	MySQLPingInterval      = 1  //in seconds
	HardBlockCheckInterval = 1  //in seconds
	MySQLTimeout           = 10 //in seconds
	RedisTimeout           = 10 //in seconds
)

const (
	StatusOk                = iota
	StatusHardblocked       = iota
	StatusRedisMasterDown   = iota
	StatusRedisSlaveDown    = iota
	StatusMySQLMasterDown   = iota
	StatusMySQLSlaveDown    = iota
	StatusRedisAndMySQLDown = iota
)

type StatusManager struct {
	run                      bool
	isHardBlocked            bool
	isTokenStorageMasterDown bool
	isTokenStorageSlaveDown  bool
	isMySQLMasterDown        bool
	isMySQLSlaveDown         bool
	forceReadOnly            bool
}

func (statusManager *StatusManager) AllowTraffic() bool {
	if statusManager.isHardBlocked {
		return false
	}

	return (!statusManager.isTokenStorageMasterDown || !statusManager.isTokenStorageSlaveDown) &&
		(!statusManager.isMySQLMasterDown || !statusManager.isMySQLSlaveDown)
}

func (statusManager *StatusManager) GetStatus() int {

	if statusManager.isHardBlocked {
		return StatusHardblocked
	}

	if (statusManager.isTokenStorageMasterDown || statusManager.isTokenStorageSlaveDown) &&
		(statusManager.isMySQLMasterDown || statusManager.isMySQLSlaveDown) {
		return StatusRedisAndMySQLDown
	}

	if statusManager.isTokenStorageMasterDown {
		return StatusRedisMasterDown
	}

	if statusManager.isTokenStorageSlaveDown {
		return StatusRedisSlaveDown
	}

	if statusManager.isMySQLMasterDown {
		return StatusMySQLMasterDown
	}

	if statusManager.isMySQLSlaveDown {
		return StatusMySQLSlaveDown
	}

	return StatusOk
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

func (statusManager *StatusManager) runIsHardBlockedCheck() {
	for statusManager.run {
		time.Sleep(HardBlockCheckInterval * time.Second)
		prevStatus := statusManager.isHardBlocked
		statusManager.isHardBlocked = statusManager.checkIsHardBlocked()
		statusManager.logStatusChange(
			"Service enabled - file block removed", "Serviced disabled due to file block", prevStatus, statusManager.isHardBlocked)
	}
}

func (statusManager *StatusManager) logStatusChange(
	availableMsg string, notAvailableMsg string, wasDown bool, isDown bool) {
	if wasDown != isDown {
		if !isDown {
			logger.GetLogger().Info(availableMsg)
		} else {
			logger.GetLogger().Error(notAvailableMsg)
		}
	}
}

func runWithTimeout(timeoutSec int, in chan error) (bool, error) {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(timeoutSec) * time.Second)
		timeout <- true
	}()

	select {
	case err := <-in:
		return false, err
	case <-timeout:
		return true, nil
	}
}

func (statusManager *StatusManager) runRedisMasterPing(redisStorage *storage.RedisStorage) {
	for statusManager.run {
		time.Sleep(RedisPingInterval * time.Second)
		prevStatus := statusManager.isTokenStorageMasterDown

		pingChan := make(chan error, 1)
		go func() {
			pingChan <- redisStorage.PingMaster()
		}()

		timeout, err := runWithTimeout(RedisTimeout, pingChan)

		statusManager.isTokenStorageMasterDown = timeout || err != nil && err.(*storage.StorageDisabledError) == nil
		statusManager.logStatusChange(
			"Redis Master Ok", "No Ping response from Redis Master", prevStatus, statusManager.isTokenStorageMasterDown)
		if !statusManager.forceReadOnly {
			redisStorage.SetForceUseSlave(statusManager.isTokenStorageMasterDown)
		}
	}
}

func (statusManager *StatusManager) runRedisSlavePing(redisStorage *storage.RedisStorage) {
	for statusManager.run {
		time.Sleep(RedisPingInterval * time.Second)
		prevStatus := statusManager.isTokenStorageSlaveDown

		pingChan := make(chan error, 1)
		go func() {
			pingChan <- redisStorage.PingSlave()
		}()

		timeout, err := runWithTimeout(RedisTimeout, pingChan)

		statusManager.isTokenStorageSlaveDown = timeout || err != nil && err.(*storage.StorageDisabledError) == nil
		statusManager.logStatusChange(
			"Redis Slave Ok", "No Ping response from Redis Slave", prevStatus, statusManager.isTokenStorageSlaveDown)
	}
}

func (statusManager *StatusManager) runMySQLMasterPing(storagePinger *models.StoragePinger) {
	for statusManager.run {
		time.Sleep(MySQLPingInterval * time.Second)
		prevStatus := statusManager.isMySQLMasterDown

		pingChan := make(chan error, 1)
		go func() {
			pingChan <- storagePinger.PingMaster()
		}()

		//The driver sometimes doesn't give a result even if a timeout is set, so we force the timeout here
		timeout, err := runWithTimeout(MySQLTimeout, pingChan)
		statusManager.isMySQLMasterDown = timeout == true || err != nil
		statusManager.logStatusChange(
			"MySQL Master Ok", "No Ping response from MySQL Master", prevStatus, statusManager.isMySQLMasterDown)
	}
}

func (statusManager *StatusManager) runMySQLSlavePing(storagePinger *models.StoragePinger) {
	for statusManager.run {
		time.Sleep(MySQLPingInterval * time.Second)
		prevStatus := statusManager.isMySQLSlaveDown

		pingChan := make(chan error, 1)
		go func() {
			pingChan <- storagePinger.PingSlave()
		}()

		//The driver sometimes doesn't give a result even if a timeout is set, so we force the timeout here
		timeout, err := runWithTimeout(MySQLTimeout, pingChan)
		statusManager.isMySQLSlaveDown = timeout == true || err != nil
		statusManager.logStatusChange(
			"MySQL Slave Ok", "No Ping response from MySQL Slave", prevStatus, statusManager.isMySQLSlaveDown)
	}
}

func NewStatusManager(serverConfig *config.ServerConfig, redisStorage *storage.RedisStorage,
	storageFactory *models.StorageFactory) *StatusManager {

	statusManager := new(StatusManager)

	statusManager.run = true
	statusManager.isHardBlocked = statusManager.checkIsHardBlocked()
	statusManager.forceReadOnly = serverConfig.ForceReadOnly
	if statusManager.forceReadOnly {
		redisStorage.SetForceUseSlave(true)
	}
	storagePinger := storageFactory.GetStoragePinger()

	go statusManager.runIsHardBlockedCheck()
	go statusManager.runRedisMasterPing(redisStorage)
	go statusManager.runRedisSlavePing(redisStorage)
	go statusManager.runMySQLMasterPing(storagePinger)
	go statusManager.runMySQLSlavePing(storagePinger)

	return statusManager
}

func (statusManager *StatusManager) Close() {
	statusManager.run = false
}
