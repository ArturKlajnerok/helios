package config

import (
	"errors"
	"fmt"
	"time"

	"code.google.com/p/gcfg"
)

type ServerConfig struct {
	Address                     string `gcfg:"address"`
	TokenExpirationInSec        int    `gcfg:"token-expiration-in-sec"`
	RefreshTokenExpirationInSec int    `gcfg:"refresh-token-expiration-in-sec"`
	AllowMultipleAccessTokens   bool   `gcfg:"allow-multiple-access-tokens"`
	ForceReadOnly               bool   `gcfg:"force-read-only"`
}

type DbConfig struct {
	ConnectionString string `gcfg:"connection-string"`
	Type             string `gcfg:"type"`
	Engine           string `gcfg:"engine"`
	Encoding         string `gcfg:"encoding"`
	UserTable        string `gcfg:"user-table"`
	UserTableKey     string `gcfg:"user-table-key"`
}

type RedisGeneralConfig struct {
	Prefix string `gcfg:"prefix"`
}

type RedisInstanceConfig struct {
	UseThisInstance bool          `gcfg:"use-this-instance"`
	Address         string        `gcfg:"address"`
	Password        string        `gcfg:"password"`
	MaxIdleConn     int           `gcfg:"max-idle-connections"`
	IdleTimeoutSec  time.Duration `gcfg:"idle-timeout-in-seconds"`
}

type Config struct {
	Server       ServerConfig        `gcfg:"server"`
	Db           DbConfig            `gcfg:"db"`
	RedisGeneral RedisGeneralConfig  `gcfg:"redis-general"`
	RedisMaster  RedisInstanceConfig `gcfg:"redis-master"`
	RedisSlave   RedisInstanceConfig `gcfg:"redis-slave"`
}

var config *Config

func LoadConfig(path string) *Config {
	var config Config
	err := gcfg.ReadFileInto(&config, path)

	if err != nil {
		msg := fmt.Sprintf("Error on config file load: %v\n", err)
		panic(errors.New(msg))
	}
	return &config
}
