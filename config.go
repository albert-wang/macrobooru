package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Endpoint      string
	UploaderEmail string
	BindAddr      string
}

func LoadConfigurationFile(path string) (Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	result := Config{}
	err = json.Unmarshal(bytes, &result)
	return result, err
}
