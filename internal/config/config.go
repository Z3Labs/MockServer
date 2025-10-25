package config

type Config struct {
	Name string
	Host string
	Port int
	Log  LogConfig
}

type LogConfig struct {
	Mode  string
	Level string
}
