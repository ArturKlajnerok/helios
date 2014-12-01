package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

type ServerConfig struct {
	Address string `json:"address"`
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
		log.Printf("ERROR: %s\n", err)
	}
	json.Unmarshal(configFile, &config)
	return config
}
