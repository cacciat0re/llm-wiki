package config

type Configs struct {
	Server ServerConfig
	DB     DatabaseConfig
}

type ServerConfig struct {
	Port           int
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
}
