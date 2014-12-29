package storage

import (
	"encoding/json"
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
	pool                        *redis.Pool
	refreshTokenExpirationInSec int
	prefix                      string
}

func NewRedisStorage(config *config.RedisConfig, serverConfig *config.ServerConfig) *RedisStorage {
	pool := newPool(config)
	return &RedisStorage{
		pool: pool,
		refreshTokenExpirationInSec: serverConfig.RefreshTokenExpirationInSec,
		prefix: config.Prefix,
	}
}

func newPool(config *config.RedisConfig) *redis.Pool {
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

//This is an inteface function called after each reponse has been handled. We do not
//want to recreate the storage object for each request, so this function is empty
func (storage *RedisStorage) Close() {}

func (storage *RedisStorage) DoClose() {
	if storage.pool != nil {
		storage.pool.Close()
		storage.pool = nil
	}
}

func (storage *RedisStorage) Clone() osin.Storage {
	return storage
}

func (storage *RedisStorage) GetClient(id string) (osin.Client, error) {
	key := createClientKey(id)
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
	key := createClientKey(id)
	clientJSON, err := json.Marshal(client)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, clientJSON)
}

func (storage *RedisStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	key := createAuthorizeKey(data.Code)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, dataJSON)
}

func (storage *RedisStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	key := createAuthorizeKey(code)
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
	key := createAuthorizeKey(code)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) SaveAccess(data *osin.AccessData) error {
	key := createAccessKey(data.AccessToken)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	err = storage.SetExpirableKey(key, dataJSON, int(data.ExpiresIn))
	if data.RefreshToken != "" {
		key_refresh := createRefreshKey(data.RefreshToken)
		err = storage.SetExpirableKey(key_refresh, dataJSON, storage.refreshTokenExpirationInSec)
	}
	if err == nil {
		err = storage.SaveAccessTokenForUserId(data)
	}
	return err
}

func (storage *RedisStorage) LoadAccess(token string) (*osin.AccessData, error) {
	key := createAccessKey(token)
	accessJSON, err := storage.GetKey(key, true)
	if err != nil {
		return nil, err
	}

	return unmarshallAccess(accessJSON)
}

func (storage *RedisStorage) RemoveAccess(token string) error {
	key := createAccessKey(token)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	key := createRefreshKey(token)
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
	key := createRefreshKey(token)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) SaveAccessTokenForUserId(data *osin.AccessData) error {
	key := createUserIdAccessKey(data.UserData.(string))
	return storage.SetExpirableKey(key, []byte(data.AccessToken), int(data.ExpiresIn))
}

func (storage *RedisStorage) GetAccessForUserId(userId string) (*osin.AccessData, error) {
	key := createUserIdAccessKey(userId)
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
	db := storage.pool.Get()
	defer db.Close()
	value, err := redis.String(db.Do("GET", keyName))
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
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("SET", key, string(value))
	logger.GetLogger().ErrorErr(err)
	return err
}

func (storage *RedisStorage) SetExpirableKey(key string, value []byte, expireInSec int) error {
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("SET", key, string(value), "EX", expireInSec)
	logger.GetLogger().ErrorErr(err)
	return err
}

func (storage *RedisStorage) DeleteKey(keyName string) error {
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("DEL", keyName)
	logger.GetLogger().ErrorErr(err)
	return err
}

func createClientKey(id string) string {
	return ClientPrefix + id
}

func createAuthorizeKey(code string) string {
	return AuthorizePrefix + code
}

func createAccessKey(token string) string {
	return AccessPrefix + token
}

func createRefreshKey(token string) string {
	return RefreshPrefix + token
}

func createUserIdAccessKey(userId string) string {
	return UserIdAccessKeyPrefix + userId
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
