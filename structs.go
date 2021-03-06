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
	Ratelimit int
	Filter FilterConfig
}

type FilterConfig struct {
	Blacklist []string
	Whitelist []string
	Sanitize []string
}

type ServerConfig struct {
	Port string
	Concurrency int
	MaxConnsPerIP int
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