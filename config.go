package main

import (
	"github.com/BurntSushi/toml"
)

var config struct {
	URL  string
	Dest string

	FilenameLength int

	Keys []key
}

type key struct {
	User string
	Pass string
}

func parseConfig() error {
	if _, err := toml.DecodeFile("./config.toml", &config); err != nil {
		return err
	}

	return nil
}