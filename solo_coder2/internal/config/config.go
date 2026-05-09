package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Redis  RedisConfig
	Etcd   EtcdConfig
	Log    LogConfig
}

type ServerConfig struct {
	Name     string
	GrpcPort int
	HTTPPort int
	WSPort   int
}

type MySQLConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	Charset      string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

type EtcdConfig struct {
	Endpoints   []string
	DialTimeout int
}

type LogConfig struct {
	Level  string
	Format string
}

var AppConfig *Config

func InitConfig(configPath string) {
	if configPath == "" {
		configPath = "./configs"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Name:     viper.GetString("server.name"),
			GrpcPort: viper.GetInt("server.grpc_port"),
			HTTPPort: viper.GetInt("server.http_port"),
			WSPort:   viper.GetInt("server.ws_port"),
		},
		MySQL: MySQLConfig{
			Host:         viper.GetString("mysql.host"),
			Port:         viper.GetInt("mysql.port"),
			Username:     viper.GetString("mysql.username"),
			Password:     viper.GetString("mysql.password"),
			Database:     viper.GetString("mysql.database"),
			Charset:      viper.GetString("mysql.charset"),
			MaxOpenConns: viper.GetInt("mysql.max_open_conns"),
			MaxIdleConns: viper.GetInt("mysql.max_idle_conns"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("redis.host"),
			Port:     viper.GetInt("redis.port"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
			PoolSize: viper.GetInt("redis.pool_size"),
		},
		Etcd: EtcdConfig{
			Endpoints:   viper.GetStringSlice("etcd.endpoints"),
			DialTimeout: viper.GetInt("etcd.dial_timeout"),
		},
		Log: LogConfig{
			Level:  viper.GetString("log.level"),
			Format: viper.GetString("log.format"),
		},
	}

	log.Printf("Config loaded successfully: %+v", AppConfig)
}

func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		m.Username, m.Password, m.Host, m.Port, m.Database, m.Charset)
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
