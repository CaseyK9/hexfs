package main

type Configuration struct {
	Security SecurityConfig
	Server ServerConfig
	Net NetConfig
}

type SecurityConfig struct {
	MasterKey string
	MaxSizeBytes int
	PublicMode bool
	Blacklist []string
	Whitelist []string
}

type ServerConfig struct {
	Port string
	IDLen int
}

type NetConfig struct {
	Redis RedisConfig
	GCS GCSConfig
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