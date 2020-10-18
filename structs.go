package main

// FileData represents a file stored in hexFS.
type FileData struct {
	ID string `json:"id,omitempty," bson:"id,omitempty"`
	Ext string `json:"ext,omitempty" bson:"ext,omitempty"`
	SHA256 string `json:"sha256,omitempty" bson:"sha256,omitempty"`
	UploadedTimestamp string `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	IP string `json:"ip,omitempty" bson:"ip,omitempty"`
	Size int64 `json:"size,omitempty" bson:"size,omitempty"`
}

type Configuration struct {
	Security SecurityConfig
	Server ServerConfig
	Net NetConfig
}

type SecurityConfig struct {
	MasterKey string
	StandardKey string
	DisableFileBlacklist bool
	MaxSizeBytes int
	Capacity int64
	PublicMode bool
}

type ServerConfig struct {
	Port string
	Frontend string
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