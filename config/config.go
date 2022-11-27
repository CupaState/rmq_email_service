package config

import (
	"errors"
	"log"
	"time"

	"github.com/spf13/viper"
)

// App config struct
type Config struct {
	Server 		ServerConfig
	RabbitMQ 	RabbitMQ
	Postgres 	PostgreSQLConfig
	Redis			RedisConfig
	Cookie		Cookie
	Session		Session
	Metrics		Metrics
	Logger 		Logger
	AWS				AWS
	Jaeger 		Jaeger
	Smtp 			Smtp
}

// Server config struct
type ServerConfig struct {
	AppVersion				string
	Port 							string
	PprofPrort 				string
	Mode 							string
	JwtSecretKey 			string
	CookieName 				string
	ReadTimeout 			time.Duration
	WriteTimeout 			time.Duration
	SSL 							bool
	CtxDefaultTimeout time.Duration
	CSRF 							bool
	Debug 						bool
	MaxConnectionIdle time.Duration
	Timeout 					time.Duration
	MaxConnectionAge  time.Duration
	Time 							time.Duration
}

// Smtp
type Smtp struct {
	Host 					string 
	Port 					int
	User 					string
	Password 			string
}

// RabbitMQ
type RabbitMQ struct {
	Host 						string
	Port 						string
	User 						string
	Password 				string
	Exchange 				string
	Queue 					string
	RoutingKey 			string
	ConsumerTag 		string
	WorkerPoolSize	int
}

// Logger config
type Logger struct {
	Development 			bool
	DisableCaller 		bool
	DisableStacktrace bool
	Encoding 					string
	Level 						string
}

// PostgreSQL config
type PostgreSQLConfig struct {
	Host						string
	Port 						string
	User 						string
	Password 				string
	DBname					string
	SSLMode 				bool
	Driver 					string
}

// Redis config
type RedisConfig struct {
	RedisAddr		 				string
	RedisPassword				string
	RedisDB							string
	RedisDefaultDb 			string
	MinIdleCons 				int
	PoolSize 						int
	PoolTimeout 				int
	Password 						string
	DB 									int
}

// Cookie config
type Cookie struct {
	Name 				string
	MaxAge 			string
	Secure 			bool
	HTTPonly 		bool
}

// Session config
type Session struct {
	Prefix 			string
	Name				string
	Expire 			int
}

// Metrics config
type Metrics struct {
	URL		 			string
	ServiceName	string
}

// AWS S3
type AWS struct {
	Endpoint 				string
	MinioAccessKey	string
	MinioSecretKey 	string
	UseSSL 					bool
	MinioEndpoint 	string
}

// Jaeger
type Jaeger struct {
	Host 				string
	ServiceName string
	LogSpans    bool
}

// Load config file from given path
func LoadConfig(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}
	return v, nil
}

// Parse config file
func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config
	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}
	
	log.Println("OLEG SOBAKA: ", c.Logger)

	return &c, nil
}

func GetConfig(confPath string) (*Config, error) {
	cfgFile, err := LoadConfig(confPath)
	//log.Println("CUPA2", cfgFile)
	if err != nil {
		return nil, err
	}

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetConfigPath(confPath string) string {
	if confPath == "docker" {
		return "./config/config-docker"
	}
	return "./config/config-local"
}
