package storage

import (
	"encoding/json"
	"github.com/RangelReale/osin"
	"github.com/Wikia/helios/config"
	"github.com/garyburd/redigo/redis"
	"time"
)

const (
	CLIENT_PREFIX    = "client."
	AUTHORIZE_PREFIX = "authorization."
	ACCESS_PREFIX    = "access."
	REFRESH_PREFIX   = "refresh."
)

type RedisStorage struct {
	pool   *redis.Pool
	prefix string
}

func NewRedisStorage(config *config.RedisConfig) *RedisStorage {
	pool := newPool(config)
	return &RedisStorage{pool: pool, prefix: config.Prefix}
}

func newPool(config *config.RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     config.MaxIdleConn,
		IdleTimeout: config.IdleTimeoutSec * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Address)
			if err != nil {
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

func (storage *RedisStorage) Close() {}

func (storage *RedisStorage) Clone() osin.Storage {
	return storage
}

func (storage *RedisStorage) GetClient(id string) (osin.Client, error) {
	key := createClientKey(id)
	clientJSON, err := storage.GetKey(key)
	if err != nil {
		return nil, err
	}

	client := new(osin.DefaultClient)
	if err := json.Unmarshal(clientJSON, &client); err != nil {
		return nil, err
	}

	return client, nil
}

func (storage *RedisStorage) SetClient(id string, client osin.Client) error {
	key := createClientKey(id)
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return err
	}

	return storage.SetKey(key, string(clientJSON))
}

func (storage *RedisStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	key := createAuthorizeKey(data.Code)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return storage.SetKey(key, string(dataJSON))
}

func (storage *RedisStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	key := createAuthorizeKey(code)
	authJSON, err := storage.GetKey(key)
	if err != nil {
		return nil, err
	}

	auth := new(osin.AuthorizeData)
	if err := json.Unmarshal(authJSON, &auth); err != nil {
		return nil, err
	}

	return auth, nil
}

func (storage *RedisStorage) RemoveAuthorize(code string) error {
	key := createAuthorizeKey(code)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) SaveAccess(data *osin.AccessData) error {
	key := createAccessKey(data.AccessToken)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = storage.SetKey(key, string(dataJSON))
	if data.RefreshToken != "" {
		key_refresh := createRefreshKey(data.RefreshToken)
		err = storage.SetKey(key_refresh, string(dataJSON))
	}
	return err
}

func (storage *RedisStorage) LoadAccess(token string) (*osin.AccessData, error) {
	key := createAccessKey(token)
	accessJSON, err := storage.GetKey(key)
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
	refreshJSON, storeErr := storage.GetKey(key)
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

func (storage *RedisStorage) GetKey(keyName string) ([]byte, error) {
	db := storage.pool.Get()
	defer db.Close()
	value, err := redis.String(db.Do("GET", keyName))
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

func (storage *RedisStorage) SetKey(key string, value string) error {
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("SET", key, value)
	if err != nil {
		return err
	}
	return nil
}

func (storage *RedisStorage) DeleteKey(keyName string) error {
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("DEL", keyName)
	if err != nil {
		return err
	}
	return nil
}

func createClientKey(id string) string {
	return CLIENT_PREFIX + id
}

func createAuthorizeKey(code string) string {
	return AUTHORIZE_PREFIX + code
}

func createAccessKey(token string) string {
	return ACCESS_PREFIX + token
}

func createRefreshKey(token string) string {
	return REFRESH_PREFIX + token
}

func unmarshallAccess(JSON []byte) (*osin.AccessData, error) {
	access := new(osin.AccessData)
	access.Client = new(osin.DefaultClient)
	access.AccessData = new(osin.AccessData)
	access.AccessData.Client = new(osin.DefaultClient)
	if err := json.Unmarshal(JSON, &access); err != nil {
		return nil, err
	}
	return access, nil
}
