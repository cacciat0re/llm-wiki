package config

type Configs struct {
	Server ServerConfig
	DB     DatabaseConfig
	CORS   CORSConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
}

type CORSConfig struct {
	AllowedOrigins []string
}
