package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// config is a stuct with all config values. See `runtime/config/config.toml`
// for more information about these values.
type config struct {
	Root string

	Listen string

	WritePath string
	ServePath string

	MaxFileSize          int64
	RandomFilenameLenght int

	AcceptedKeys []key
}

type key struct {
	User string
	Pass string
}

// parseConfig parses a toml config.
func parseConfig() (*config, error) {
	c := &config{}

	if _, err := toml.DecodeFile("/etc/punfed/upload.toml", c); err != nil {
		return nil, fmt.Errorf("config %s: %s", "/etc/punfed/upload.toml", err)
	}

	// Convert `MaxFileSize` to MB.
	c.MaxFileSize = c.MaxFileSize << 20

	return c, nil
}
