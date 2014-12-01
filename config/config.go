package config

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
	Address  string `json:"address"`
	Password string `json:"password"`
	Prefix   string `json:"prefix"`
}

type Config struct {
	Server *ServerConfig `json:"server"`
	Db     *DbConfig     `json:"db"`
	Redis  *RedisConfig  `json:"redis"`
}

var config *Config

func LoadConfig() *Config {
	server := new(ServerConfig)
	server.Address = ":8080"

	db := new(DbConfig)
	db.Type = "mysql"
	db.Engine = "InnoDB"
	db.Encoding = "UTF8"
	db.UserTable = "user"
	db.UserTableKey = "Id"

	redis := new(RedisConfig)
	redis.Address = "localhost:6379"
	redis.Password = ""
	redis.Prefix = "auth"

	config := new(Config)
	config.Server = server
	config.Db = db
	config.Redis = redis
	return config
}
