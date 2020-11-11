package main

type Configuration struct {
	Security SecurityConfig
	Server ServerConfig
	Net NetConfig
}

type SecurityConfig struct {
	MasterKey string
	MaxSizeBytes int
	Capacity int64
	PublicMode bool
	Blacklist []string
	Whitelist []string
}

type ServerConfig struct {
	Port string
	IDLen int
}

type NetConfig struct {
	Mongo MongoConfig
	Redis RedisConfig
	GCS GCSConfig
}

type MongoConfig struct {
	URI string
	Database string
}

type RedisConfig struct {
	URI string
	Password string
	Db int
}

type GCSConfig struct {
	BucketName string
	SecretKey string
}