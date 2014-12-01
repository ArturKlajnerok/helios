package config

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
	Db    *DbConfig    `json:"db"`
	Redis *RedisConfig `json:"redis"`
}

var config *Config

func LoadConfig() *Config {
	dbConfig := new(DbConfig)
	dbConfig.Type = "mysql"
	dbConfig.Engine = "InnoDB"
	dbConfig.Encoding = "UTF8"
	dbConfig.UserTable = "user"
	dbConfig.UserTableKey = "Id"

	redisConfig := new(RedisConfig)
	redisConfig.Address = "localhost:6379"
	redisConfig.Password = ""
	redisConfig.Prefix = "auth"

	config := new(Config)
	config.Db = dbConfig
	config.Redis = redisConfig
	return config
}
