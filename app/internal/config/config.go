package config

var Configuration Config

type Config struct {
	Directory string
}

func (c Config) GetDirectory() string {

	return c.Directory
}

func (c Config) SetDirectory(dir string) {
	c.Directory = dir
}
