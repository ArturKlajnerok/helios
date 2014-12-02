package storage

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/RangelReale/osin"
	"github.com/Wikia/go-commons/logger"
	"github.com/Wikia/helios/config"
	"github.com/garyburd/redigo/redis"
)

const (
	CLIENT_PREFIX    = "client."
	AUTHORIZE_PREFIX = "authorization."
	ACCESS_PREFIX    = "access."
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

func (storage *RedisStorage) Close() {}

func (storage *RedisStorage) Clone() osin.Storage {
	return storage
}

func (storage *RedisStorage) GetClient(id string) (osin.Client, error) {
	key := CreateClientKey(id)
	clientJSON, err := storage.GetKey(key)
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
	key := CreateClientKey(id)
	clientJSON, err := json.Marshal(client)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, string(clientJSON))
}

func (storage *RedisStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	key := CreateAuthorizeKey(data.Code)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, string(dataJSON))
}

func (storage *RedisStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	key := CreateAuthorizeKey(code)
	authJSON, err := storage.GetKey(key)
	if err != nil {
		return nil, err
	}

	auth := new(osin.AuthorizeData)
	if err := json.Unmarshal(authJSON, &auth); err != nil {
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}

	return auth, nil
}

func (storage *RedisStorage) RemoveAuthorize(code string) error {
	key := CreateAuthorizeKey(code)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) SaveAccess(data *osin.AccessData) error {
	key := CreateAccessKey(data.AccessToken)
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}

	return storage.SetKey(key, string(dataJSON))
}

func (storage *RedisStorage) LoadAccess(token string) (*osin.AccessData, error) {
	key := CreateAccessKey(token)
	accessJSON, err := storage.GetKey(key)
	if err != nil {
		return nil, err
	}

	access := new(osin.AccessData)
	if err := json.Unmarshal(accessJSON, &access); err != nil {
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}

	return access, nil
}

func (storage *RedisStorage) RemoveAccess(token string) error {
	key := CreateAccessKey(token)
	return storage.DeleteKey(key)
}

func (storage *RedisStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	return nil, errors.New("Not implemented")
}

func (storage *RedisStorage) RemoveRefresh(token string) error {
	return errors.New("Not implemented")
}

func (storage *RedisStorage) GetKey(keyName string) ([]byte, error) {
	db := storage.pool.Get()
	defer db.Close()
	value, err := redis.String(db.Do("GET", keyName))
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return nil, err
	}
	return []byte(value), nil
}

func (storage *RedisStorage) SetKey(key string, value string) error {
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("SET", key, value)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}
	return nil
}

func (storage *RedisStorage) DeleteKey(keyName string) error {
	db := storage.pool.Get()
	defer db.Close()
	_, err := db.Do("DEL", keyName)
	if err != nil {
		logger.GetLogger().ErrorErr(err)
		return err
	}
	return nil
}

func CreateClientKey(id string) string {
	return CLIENT_PREFIX + id
}

func CreateAuthorizeKey(code string) string {
	return AUTHORIZE_PREFIX + code
}

func CreateAccessKey(token string) string {
	return ACCESS_PREFIX + token
}
