package main

import (
	"github.com/cwc/webconf"
)

type Config struct {
	Endpoint      string `name:"Endpoint" desc:"The url of a nodebooru instance, like http://nodebooru.example.com`
	UploaderEmail string `name:"Uploader Email" desc:"An authorized email to upload as"`
	BindAddr      string `name:"Bind address" desc:"An address on which the macrobooru web service will listen"`
}

func LoadConfigurationFile(path string) (Config, error) {
	cfg := Config{}
	if c := webconf.New(&cfg, "config.json", ":50000", nil); c != nil {
		err := c.Load()
		if err != nil {
			return cfg, err
		}

		return cfg, nil
	}

	return cfg, nil
}
