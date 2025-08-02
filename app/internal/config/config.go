package config

var Configuration Config

type Config struct {
	directory string
}

func (c Config) GetDirectory() string {
	return c.directory
}

func (c *Config) SetDirectory(dir string) {
	c.directory = dir
}
