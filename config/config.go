package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type ServerConfig struct {
	Address                     string `json:"address"`
	TokenExpirationInSec        int    `json:"token_expiration_in_sec"`
	RefreshTokenExpirationInSec int    `json:"refresh_token_expiration_in_sec"`
	AllowMultipleAccessTokens   bool   `json:"allow_multiple_access_tokens"`
}

type DbConfig struct {
	Type         string `json:"type"`
	Engine       string `json:"engine"`
	Encoding     string `json:"encoding"`
	UserTable    string `json:"user_table"`
	UserTableKey string `json:"user_table_key"`
}

type RedisConfig struct {
	Address        string        `json:"address"`
	Password       string        `json:"password"`
	Prefix         string        `json:"prefix"`
	MaxIdleConn    int           `json:"max_idle_connections"`
	IdleTimeoutSec time.Duration `json:"idle_timeout_in_seconds"`
}

type Config struct {
	Server *ServerConfig `json:"server"`
	Db     *DbConfig     `json:"db"`
	Redis  *RedisConfig  `json:"redis"`
}

var config *Config

func LoadConfig(path string) *Config {
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error on config file load: %s\n", err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Printf("Error on config unmarshal: %s\n", err)
	}
	return config
}
