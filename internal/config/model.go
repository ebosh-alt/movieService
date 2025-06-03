package config

type Config struct {
	Server   ServerConfig   `yaml:"Server"`
	Postgres PostgresConfig `yaml:"Postgres"`
	Secret   string         `yaml:"Secret"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"DBName"`
	SSLMode  string `yaml:"sslMode"`
	DSN      string `yaml:"-"` // "-" означает, что это поле не будет загружаться из YAML
}

type ServerConfig struct {
	AppVersion string `yaml:"appVersion"`
	Host       string `yaml:"host" validate:"required"`
	Port       string `yaml:"port" validate:"required"`
	HTTPPort   string `yaml:"httpPort"`
}
