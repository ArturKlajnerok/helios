package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/config"
	"github.com/garyburd/redigo/redis"
)

const (
	ClientPrefix          = "client."
	AuthorizePrefix       = "authorization."
	AccessPrefix          = "access."
	RefreshPrefix         = "refresh."
	UserIdAccessKeyPrefix = "userIdAccessKey."
)

type RedisStorage struct {
	masterPool                  *redis.Pool
	slavePool                   *redis.Pool
	forceUseSlave               bool
	refreshTokenExpirationInSec int
	prefix                      string
}

type StorageDisabledError struct {
}

func (e *StorageDisabledError) Error() string { return "The given Redis Storage is not available" }

func NewRedisStorage(
	generalConfig *config.RedisGeneralConfig,
	masterConfig *config.RedisInstanceConfig,
	slaveConfig *config.RedisInstanceConfig,
	serverConfig *config.ServerConfig) *RedisStorage {

	var masterPool *redis.Pool
	if masterConfig.UseThisInstance {
		masterPool = newPool(masterConfig)
	}
	var slavePool *redis.Pool
	if slaveConfig.UseThisInstance {
		slavePool = newPool(slaveConfig)
	}
	if masterPool == nil && slavePool == nil {
		panic(errors.New("Neither Redis master pool nor slave have been configured"))
	}
	return &RedisStorage{
		masterPool:                  masterPool,
		slavePool:                   slavePool,
		refreshTokenExpirationInSec: serverConfig.RefreshTokenExpirationInSec,
		forceUseSlave:               false,
		prefix:                      generalConfig.Prefix,
	}
}

func newPool(config *config.RedisInstanceConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     config.MaxIdleConn,
		IdleTimeout: config.IdleTimeoutSec * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Address)
			if err != nil {
				logger.GetLogger().ErrorErr(err)
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (storage *RedisStorage) SetForceUseSlave(forceUseSlave bool) {
	storage.forceUseSlave = forceUseSlave
}

//This is an inteface function called after each reponse has been handled. We do not
//want to recreate the storage object for each request, so this function is empty
func (storage *RedisStorage) Close() {}

func (storage *RedisStorage) DoClose() {
	if storage.masterPool != nil {
		storage.masterPool.Close()
		storage.masterPool = nil
	}
	if storage.slavePool != nil {
		storage.slavePool.Close()
		storage.slavePool = nil
	}
}

func (storage *RedisStorage) Clone() osin.Storage {
	return storage
}

func (storage *RedisStorage) GetClient(id string) (osin.Client, error) {
	key := storage.createClientKey(id)
	clientJSON, err := storage.GetKey(key, true)
	if err != nil {
		return nil, err
	}

	client := new(osin.DefaultClient)
	if err := json.Unmarshal(clientJSON, &client); err != nil {
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}

	return client, nil
}

func (storage *RedisStorage) SetClient(id string, client osin.Client) error {
	key := storage.createClientKey(id)
	clientJSON, err := json.Marshal(client)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, clientJSON)
}

func (storage *RedisStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	key := storage.createAuthorizeKey(data.Code)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, dataJSON)
}

func (storage *RedisStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	key := storage.createAuthorizeKey(code)
	authJSON, err := storage.GetKey(key, true)
	if err != nil {
		return nil, err
	}

	auth := new(osin.AuthorizeData)
	err = json.Unmarshal(authJSON, &auth)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		auth = nil
	}
	return auth, err
}

func (storage *RedisStorage) RemoveAuthorize(code string) error {
	key := storage.createAuthorizeKey(code)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) SaveAccess(data *osin.AccessData) error {
	key := storage.createAccessKey(data.AccessToken)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	err = storage.SetExpirableKey(key, dataJSON, int(data.ExpiresIn))
	if data.RefreshToken != "" {
		key_refresh := storage.createRefreshKey(data.RefreshToken)
		err = storage.SetExpirableKey(key_refresh, dataJSON, storage.refreshTokenExpirationInSec)
	}
	if err == nil {
		err = storage.SaveAccessTokenForUserId(data)
	}
	return err
}

func (storage *RedisStorage) LoadAccess(token string) (*osin.AccessData, error) {
	key := storage.createAccessKey(token)
	accessJSON, err := storage.GetKey(key, true)
	if err != nil {
		return nil, err
	}

	return unmarshallAccess(accessJSON)
}

func (storage *RedisStorage) RemoveAccess(token string) error {
	key := storage.createAccessKey(token)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	key := storage.createRefreshKey(token)
	refreshJSON, storeErr := storage.GetKey(key, true)
	if storeErr != nil {
		return nil, storeErr
	}

	access, err := unmarshallAccess(refreshJSON)
	if err != nil {
		return nil, err
	}

	// Clear old access data
	if access.AccessData != nil {
		access.AccessData = nil
	}
	return access, nil
}

func (storage *RedisStorage) RemoveRefresh(token string) error {
	key := storage.createRefreshKey(token)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) SaveAccessTokenForUserId(data *osin.AccessData) error {
	key := storage.createUserIdAccessKey(data.UserData.(string))
	return storage.SetExpirableKey(key, []byte(data.AccessToken), int(data.ExpiresIn))
}

func (storage *RedisStorage) GetAccessForUserId(userId string) (*osin.AccessData, error) {
	key := storage.createUserIdAccessKey(userId)
	accessToken, err := storage.GetKey(key, false)
	if err != nil {
		return nil, err
	}
	if accessToken == nil {
		return nil, nil
	}
	return storage.LoadAccess(string(accessToken))
}

func (storage *RedisStorage) GetKey(keyName string, mustExist bool) ([]byte, error) {
	pool, err := storage.getPoolForRead()
	if err != nil {
		return nil, err
	}

	db := pool.Get()
	defer db.Close()

	var value string
	value, err = redis.String(db.Do("GET", keyName))
	if err != nil {
		if mustExist || err != redis.ErrNil {
			logger.GetLogger().Error(fmt.Sprintf("Error while getting key %s: %s", keyName, err.Error()))
			return nil, err
		} else {
			return nil, nil
		}

	}
	return []byte(value), nil
}

func (storage *RedisStorage) SetKey(key string, value []byte) error {
	pool, err := storage.getPoolForRead()
	if err != nil {
		return err
	}

	db := pool.Get()
	defer db.Close()
	_, err = db.Do("SET", key, string(value))
	logger.GetLogger().ErrorErr(err)
	return err
}

func (storage *RedisStorage) SetExpirableKey(key string, value []byte, expireInSec int) error {
	pool, err := storage.getPoolForRead()
	if err != nil {
		return err
	}

	db := pool.Get()
	defer db.Close()
	_, err = db.Do("SET", key, string(value), "EX", expireInSec)
	logger.GetLogger().ErrorErr(err)
	return err
}

func (storage *RedisStorage) DeleteKey(keyName string) error {
	pool, err := storage.getPoolForRead()
	if err != nil {
		return err
	}

	db := pool.Get()
	defer db.Close()
	_, err = db.Do("DEL", keyName)
	logger.GetLogger().ErrorErr(err)
	return err
}

func (storage *RedisStorage) getPoolForWrite() (*redis.Pool, error) {
	if storage.forceUseSlave {
		err := errors.New("Use slave flag is on, cannot get redis pool for writing")
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}

	if storage.masterPool == nil {
		err := errors.New("Master pool has not been configured, cannot get redis pool for writing")
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}

	return storage.masterPool, nil
}

func (storage *RedisStorage) getPoolForRead() (*redis.Pool, error) {
	if storage.forceUseSlave && storage.slavePool == nil {
		err := errors.New("Use slave flag is on, but slave pool has not been configured, cannot get redis pool for reading")
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}

	if storage.masterPool == nil || storage.forceUseSlave {
		return storage.slavePool, nil
	} else {
		return storage.masterPool, nil
	}
}

func (storage *RedisStorage) PingMaster() error {
	return storage.Ping(storage.masterPool)
}

func (storage *RedisStorage) PingSlave() error {
	return storage.Ping(storage.slavePool)
}

func (storage *RedisStorage) Ping(pool *redis.Pool) error {
	if pool == nil {
		return new(StorageDisabledError)
	}
	db := pool.Get()
	defer db.Close()

	_, err := db.Do("PING")

	return err
}

func (storage *RedisStorage) createClientKey(id string) string {
	return storage.prefix + ClientPrefix + id
}

func (storage *RedisStorage) createAuthorizeKey(code string) string {
	return storage.prefix + AuthorizePrefix + code
}

func (storage *RedisStorage) createAccessKey(token string) string {
	return storage.prefix + AccessPrefix + token
}

func (storage *RedisStorage) createRefreshKey(token string) string {
	return storage.prefix + RefreshPrefix + token
}

func (storage *RedisStorage) createUserIdAccessKey(userId string) string {
	return storage.prefix + UserIdAccessKeyPrefix + userId
}

func unmarshallAccess(JSON []byte) (*osin.AccessData, error) {
	access := new(osin.AccessData)
	access.Client = new(osin.DefaultClient)
	access.AccessData = new(osin.AccessData)
	access.AccessData.Client = new(osin.DefaultClient)
	err := json.Unmarshal(JSON, &access)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		access = nil
	}
	return access, err
}
