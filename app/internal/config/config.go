package config

type Config struct {
	Directory string
}

var Instance = &Config{}
